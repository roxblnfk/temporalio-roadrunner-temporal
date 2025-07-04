package aggregatedpool

import (
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/roadrunner-server/errors"
	"github.com/roadrunner-server/goridge/v3/pkg/frame"
	"github.com/roadrunner-server/pool/payload"
	"github.com/temporalio/roadrunner-temporal/v5/internal"
	commonpb "go.temporal.io/api/common/v1"
	bindings "go.temporal.io/sdk/internalbindings"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

const (
	completed string = "completed"
	// update types
	valExec string = "validate_execute"
)

// execution context.
func (wp *Workflow) getContext() *internal.Context {
	return &internal.Context{
		TaskQueue:              wp.env.WorkflowInfo().TaskQueueName,
		TickTime:               wp.env.Now().Format(time.RFC3339),
		Replay:                 wp.env.IsReplaying(),
		HistoryLen:             wp.env.WorkflowInfo().GetCurrentHistoryLength(),
		HistorySize:            wp.env.WorkflowInfo().GetCurrentHistorySize(),
		ContinueAsNewSuggested: wp.env.WorkflowInfo().GetContinueAsNewSuggested(),
		RrID:                   wp.rrID,
		WorkerPID: 			    wp.WorkerPID,
	}
}

func (wp *Workflow) handleUpdate(name string, id string, input *commonpb.Payloads, header *commonpb.Header, callbacks bindings.UpdateCallbacks) {
	wp.log.Debug("update request received", zap.String("RunID", wp.env.WorkflowInfo().WorkflowExecution.RunID), zap.String("name", name), zap.String("id", id))

	// save update name
	wp.updatesQueue[name] = struct{}{}
	rid := wp.env.WorkflowInfo().WorkflowExecution.RunID

	// this callback executed in the OnTick function
	updatesQueueCb := func() {
		// validate callback
		wp.updateValidateCb[id] = func(msg *internal.Message) {
			wp.log.Debug("validate request callback", zap.String("RunID", wp.env.WorkflowInfo().WorkflowExecution.RunID), zap.String("name", name), zap.String("id", id), zap.Bool("is_replaying", wp.env.IsReplaying()), zap.Any("result", msg))
			if !wp.env.IsReplaying() {
				// before acceptance, we have only one option - reject
				if msg.Failure != nil {
					callbacks.Reject(temporal.GetDefaultFailureConverter().FailureToError(msg.Failure))
					return
				}
			}

			// update should be accepted on validating
			callbacks.Accept()
		}

		// execute callback
		wp.updateCompleteCb[id] = func(msg *internal.Message) {
			wp.log.Debug("update request callback", zap.String("RunID", wp.env.WorkflowInfo().WorkflowExecution.RunID), zap.String("name", name), zap.String("id", id), zap.Any("result", msg))
			if msg.Failure != nil {
				callbacks.Complete(nil, temporal.GetDefaultFailureConverter().FailureToError(msg.Failure))
				return
			}

			callbacks.Complete(msg.Payloads, nil)
		}

		// push validate command
		wp.mq.PushCommand(
			&internal.InvokeUpdate{
				RunID:    rid,
				UpdateID: id,
				Name:     name,
				Type:     valExec,
			},
			input,
			header,
		)
	}

	wp.env.QueueUpdate(name, updatesQueueCb)
}

// schedule cancel command
func (wp *Workflow) handleCancel() {
	wp.mq.PushCommand(
		internal.CancelWorkflow{RunID: wp.env.WorkflowInfo().WorkflowExecution.RunID},
		nil,
		wp.header,
	)
}

// schedule the signal processing
func (wp *Workflow) handleSignal(name string, input *commonpb.Payloads, header *commonpb.Header) error {
	wp.log.Debug("signal request", zap.String("RunID", wp.env.WorkflowInfo().WorkflowExecution.RunID), zap.String("name", name))
	wp.mq.PushCommand(
		internal.InvokeSignal{
			RunID: wp.env.WorkflowInfo().WorkflowExecution.RunID,
			Name:  name,
		},
		input,
		header,
	)

	return nil
}

// Handle query in blocking mode.
func (wp *Workflow) handleQuery(queryType string, queryArgs *commonpb.Payloads, header *commonpb.Header) (*commonpb.Payloads, error) {
	const op = errors.Op("workflow_process_handle_query")

	wp.log.Debug("query request", zap.String("RunID", wp.env.WorkflowInfo().WorkflowExecution.RunID), zap.String("name", queryType))

	result, err := wp.runCommand(internal.InvokeQuery{
		RunID: wp.env.WorkflowInfo().WorkflowExecution.RunID,
		Name:  queryType,
	}, queryArgs, header)

	if err != nil {
		return nil, errors.E(op, err)
	}

	if result.Failure != nil {
		return nil, errors.E(op, temporal.GetDefaultFailureConverter().FailureToError(result.Failure))
	}

	return result.Payloads, nil
}

// Workflow incoming command
func (wp *Workflow) handleMessage(msg *internal.Message) error {
	const op = errors.Op("handleMessage")

	switch command := msg.Command.(type) {
	case *internal.ExecuteActivity:
		wp.log.Debug("activity request", zap.Uint64("ID", msg.ID))
		params := command.ActivityParams(wp.env, msg.Payloads, msg.Header)
		activityID := wp.env.ExecuteActivity(params, wp.createCallback(msg.ID, "activity"))

		wp.canceller.Register(msg.ID, func() error {
			wp.log.Debug("registering activity canceller", zap.String("activityID", activityID.String()))
			wp.env.RequestCancelActivity(activityID)
			return nil
		})

	case *internal.ExecuteLocalActivity:
		wp.log.Debug("local activity request", zap.Uint64("ID", msg.ID))
		params := command.LocalActivityParams(wp.env, wp.la, msg.Payloads, msg.Header)
		activityID := wp.env.ExecuteLocalActivity(params, wp.createLocalActivityCallback(msg.ID))
		wp.canceller.Register(msg.ID, func() error {
			wp.log.Debug("registering local activity canceller", zap.String("activityID", activityID.String()))
			wp.env.RequestCancelLocalActivity(activityID)
			return nil
		})

	case *internal.ExecuteChildWorkflow:
		wp.log.Debug("execute child workflow request", zap.Uint64("ID", msg.ID))
		params := command.WorkflowParams(wp.env, msg.Payloads, msg.Header)

		// always use deterministic id
		if params.WorkflowID == "" {
			nextID := atomic.AddUint64(&wp.seqID, 1)
			params.WorkflowID = wp.env.WorkflowInfo().WorkflowExecution.RunID + "_" + strconv.Itoa(int(nextID)) //nolint:gosec
		}

		wp.env.ExecuteChildWorkflow(params, wp.createCallback(msg.ID, "ExecuteChildWorkflow"), func(r bindings.WorkflowExecution, e error) {
			wp.ids.Push(msg.ID, r, e)
		})

		wp.canceller.Register(msg.ID, func() error {
			wp.env.RequestCancelChildWorkflow(params.Namespace, params.WorkflowID)
			return nil
		})

	case *internal.GetChildWorkflowExecution:
		wp.log.Debug("get child workflow execution request", zap.Uint64("ID", msg.ID))
		wp.ids.Listen(command.ID, func(w bindings.WorkflowExecution, err error) {
			cl := wp.createCallback(msg.ID, "GetChildWorkflow")

			if err != nil {
				cl(nil, err)
				return
			}

			p, er := wp.env.GetDataConverter().ToPayloads(w)
			if er != nil {
				panic(er)
			}

			cl(p, err)
		})

	case *internal.NewTimer:
		wp.log.Debug("timer request", zap.Uint64("ID", msg.ID))
		timerID := wp.env.NewTimer(command.ToDuration(), workflow.TimerOptions{
			Summary: command.Summary,
		}, wp.createCallback(msg.ID, "NewTimer"))
		wp.canceller.Register(msg.ID, func() error {
			if timerID != nil {
				wp.log.Debug("cancel timer request", zap.String("timerID", timerID.String()))
				wp.env.RequestCancelTimer(*timerID)
			}
			return nil
		})

	case *internal.GetVersion:
		wp.log.Debug("get version request", zap.Uint64("ID", msg.ID))
		version := wp.env.GetVersion(
			command.ChangeID,
			workflow.Version(command.MinSupported),
			workflow.Version(command.MaxSupported),
		)

		result, err := wp.env.GetDataConverter().ToPayloads(version)
		if err != nil {
			return errors.E(op, err)
		}

		wp.mq.PushResponse(msg.ID, result)
		err = wp.flushQueue()
		if err != nil {
			return errors.E(op, err)
		}

	case *internal.SideEffect:
		wp.log.Debug("side-effect request", zap.Uint64("ID", msg.ID))
		wp.env.SideEffect(
			func() (*commonpb.Payloads, error) {
				return msg.Payloads, nil
			},
			wp.createContinuableCallback(msg.ID, "SideEffect"),
		)

	case *internal.UpdateCompleted:
		wp.log.Debug("complete update request", zap.String("update id", command.ID))

		if command.ID == "" {
			wp.log.Error("update id is empty, can't complete update", zap.String("workflow id", wp.env.WorkflowInfo().WorkflowExecution.ID), zap.String("run id", wp.env.WorkflowInfo().WorkflowExecution.RunID))
			return errors.Str("update id is empty, can't complete update")
		}

		if _, ok := wp.updateCompleteCb[command.ID]; !ok {
			wp.log.Warn("no such update ID, can't complete update", zap.String("requested id", command.ID))
			// TODO(rustatian): error here?
			return nil
		}

		wp.updateCompleteCb[command.ID](msg)
		delete(wp.updateCompleteCb, command.ID)

	case *internal.UpdateValidated:
		wp.log.Debug("validate update request", zap.String("update id", command.ID))

		if command.ID == "" {
			wp.log.Error("update id is empty, can't validate update", zap.String("workflow id", wp.env.WorkflowInfo().WorkflowExecution.ID), zap.String("run id", wp.env.WorkflowInfo().WorkflowExecution.RunID))
			return errors.Str("update id is empty, can't validate update")
		}

		if _, ok := wp.updateValidateCb[command.ID]; !ok {
			wp.log.Warn("no such update ID, can't validate update", zap.String("requested id", command.ID))
			// TODO(rustatian): error here?
			return nil
		}

		wp.updateValidateCb[command.ID](msg)
		delete(wp.updateValidateCb, command.ID)
		// delete updateCompleteCb in case of error
		if msg.Failure != nil {
			delete(wp.updateCompleteCb, command.ID)
		}

	case *internal.CompleteWorkflow:
		wp.log.Debug("complete workflow request", zap.Uint64("ID", msg.ID))
		result, _ := wp.env.GetDataConverter().ToPayloads(completed)
		wp.mq.PushResponse(msg.ID, result)

		if msg.Failure == nil {
			wp.env.Complete(msg.Payloads, nil)
			return nil
		}

		wp.env.Complete(nil, temporal.GetDefaultFailureConverter().FailureToError(msg.Failure))

	case *internal.ContinueAsNew:
		wp.log.Debug("continue-as-new request", zap.Uint64("ID", msg.ID), zap.String("name", command.Name))
		result, _ := wp.env.GetDataConverter().ToPayloads(completed)
		wp.mq.PushResponse(msg.ID, result)

		wp.env.Complete(nil, &workflow.ContinueAsNewError{
			WorkflowType: &bindings.WorkflowType{
				Name: command.Name,
			},
			Input:               msg.Payloads,
			Header:              msg.Header,
			TaskQueueName:       command.Options.TaskQueueName,
			WorkflowRunTimeout:  command.Options.WorkflowRunTimeout,
			WorkflowTaskTimeout: command.Options.WorkflowTaskTimeout,
		})

	case *internal.UpsertWorkflowSearchAttributes:
		wp.log.Debug("upsert search attributes request", zap.Uint64("ID", msg.ID))
		err := wp.env.UpsertSearchAttributes(command.SearchAttributes)
		if err != nil {
			return errors.E(op, err)
		}

	case *internal.UpsertWorkflowTypedSearchAttributes:
		wp.log.Debug("upsert typed search attributes request", zap.Uint64("ID", msg.ID), zap.Any("search_attributes", command.SearchAttributes))
		var sau []temporal.SearchAttributeUpdate

		for k, v := range command.SearchAttributes {
			switch v.Type {
			case internal.BoolType:
				if v.Operation == internal.TypedSearchAttributeOperationUnset {
					sau = append(sau, temporal.NewSearchAttributeKeyBool(k).ValueUnset())
					continue
				}
				if v.Value == nil {
					wp.log.Warn("field value is not set", zap.String("key", k))
					continue
				}

				if tt, ok := v.Value.(bool); ok {
					sau = append(sau, temporal.NewSearchAttributeKeyBool(k).ValueSet(tt))
				} else {
					wp.log.Warn("field value is not a bool type", zap.String("key", k), zap.Any("value", v.Value))
				}

			case internal.FloatType:
				if v.Operation == internal.TypedSearchAttributeOperationUnset {
					sau = append(sau, temporal.NewSearchAttributeKeyFloat64(k).ValueUnset())
					continue
				}

				if v.Value == nil {
					wp.log.Warn("field value is not set", zap.String("key", k))
					continue
				}

				if tt, ok := v.Value.(float64); ok {
					sau = append(sau, temporal.NewSearchAttributeKeyFloat64(k).ValueSet(tt))
				} else {
					wp.log.Warn("field value is not a float64 type", zap.String("key", k), zap.Any("value", v.Value))
				}

			case internal.IntType:
				if v.Operation == internal.TypedSearchAttributeOperationUnset {
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueUnset())
					continue
				}

				if v.Value == nil {
					wp.log.Warn("field value is not set", zap.String("key", k))
					continue
				}

				switch ti := v.Value.(type) {
				case float64:
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueSet(int64(ti)))
				case int:
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueSet(int64(ti)))
				case int64:
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueSet(ti))
				case int32:
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueSet(int64(ti)))
				case int16:
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueSet(int64(ti)))
				case int8:
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueSet(int64(ti)))
				case string:
					i, err := strconv.ParseInt(ti, 10, 64)
					if err != nil {
						wp.log.Warn("failed to parse int", zap.Error(err))
						continue
					}
					sau = append(sau, temporal.NewSearchAttributeKeyInt64(k).ValueSet(i))
				default:
					wp.log.Warn("field value is not an int type", zap.String("key", k), zap.Any("value", v.Value))
				}

			case internal.KeywordType:
				if v.Operation == internal.TypedSearchAttributeOperationUnset {
					sau = append(sau, temporal.NewSearchAttributeKeyKeyword(k).ValueUnset())
					continue
				}

				if v.Value == nil {
					wp.log.Warn("field value is not set", zap.String("key", k))
					continue
				}

				if tt, ok := v.Value.(string); ok {
					sau = append(sau, temporal.NewSearchAttributeKeyKeyword(k).ValueSet(tt))
				} else {
					wp.log.Warn("field value is not a string type", zap.String("key", k), zap.Any("value", v.Value))
				}
			case internal.KeywordListType:
				if v.Operation == internal.TypedSearchAttributeOperationUnset {
					sau = append(sau, temporal.NewSearchAttributeKeyKeywordList(k).ValueUnset())
					continue
				}

				if v.Value == nil {
					wp.log.Warn("field value is not set", zap.String("key", k))
					continue
				}

				switch tt := v.Value.(type) {
				case []string:
					sau = append(sau, temporal.NewSearchAttributeKeyKeywordList(k).ValueSet(tt))
				case []any:
					var res []string
					for _, v := range tt {
						if s, ok := v.(string); ok {
							res = append(res, s)
						}
					}
					sau = append(sau, temporal.NewSearchAttributeKeyKeywordList(k).ValueSet(res))
				default:
					wp.log.Warn("field value is not a []string (strings array) type", zap.String("key", k), zap.Any("value", v.Value))
				}

			case internal.StringType:
				if v.Operation == internal.TypedSearchAttributeOperationUnset {
					sau = append(sau, temporal.NewSearchAttributeKeyString(k).ValueUnset())
					continue
				}

				if v.Value == nil {
					wp.log.Warn("field value is not set", zap.String("key", k))
					continue
				}

				if tt, ok := v.Value.(string); ok {
					sau = append(sau, temporal.NewSearchAttributeKeyString(k).ValueSet(tt))
				} else {
					wp.log.Warn("field value is not a string type", zap.String("key", k), zap.Any("value", v.Value))
				}
			case internal.DatetimeType:
				if v.Operation == internal.TypedSearchAttributeOperationUnset {
					sau = append(sau, temporal.NewSearchAttributeKeyTime(k).ValueUnset())
					continue
				}

				if v.Value == nil {
					wp.log.Warn("field value is not set", zap.String("key", k))
					continue
				}

				if tt, ok := v.Value.(string); ok {
					tm, err := time.Parse(time.RFC3339, tt)
					if err != nil {
						return errors.E(op, fmt.Errorf("failed to parse time into RFC3339: %w", err))
					}

					sau = append(sau, temporal.NewSearchAttributeKeyTime(k).ValueSet(tm))
				} else {
					wp.log.Warn("bool field value is not a bool type", zap.String("key", k), zap.Any("value", v.Value))
				}
			}
		}

		if len(sau) == 0 {
			wp.log.Warn("search attributes called, but no attributes were set")
			return nil
		}

		err := wp.env.UpsertTypedSearchAttributes(temporal.NewSearchAttributes(sau...))
		if err != nil {
			return errors.E(op, err)
		}

	case *internal.SignalExternalWorkflow:
		wp.log.Debug("signal external workflow request", zap.Uint64("ID", msg.ID))
		wp.env.SignalExternalWorkflow(
			command.Namespace,
			command.WorkflowID,
			command.RunID,
			command.Signal,
			msg.Payloads,
			nil,
			msg.Header,
			command.ChildWorkflowOnly,
			wp.createCallback(msg.ID, "SignalExternalWorkflow"),
		)

	case *internal.CancelExternalWorkflow:
		wp.log.Debug("cancel external workflow request", zap.Uint64("ID", msg.ID))
		wp.env.RequestCancelExternalWorkflow(command.Namespace, command.WorkflowID, command.RunID, wp.createCallback(msg.ID, "CancelExternalWorkflow"))

	case *internal.Cancel:
		wp.log.Debug("cancel request", zap.Uint64("ID", msg.ID))
		err := wp.canceller.Cancel(command.CommandIDs...)
		if err != nil {
			return errors.E(op, err)
		}

		result, _ := wp.env.GetDataConverter().ToPayloads(completed)
		wp.mq.PushResponse(msg.ID, result)

		err = wp.flushQueue()
		if err != nil {
			return errors.E(op, err)
		}

	case *internal.Panic:
		wp.log.Debug("panic", zap.String("failure", msg.Failure.String()))
		// do not wrap error to pass it directly to Temporal
		return temporal.GetDefaultFailureConverter().FailureToError(msg.Failure)

	case *internal.UpsertMemo:
		wp.log.Debug("upsert memo request", zap.Uint64("ID", msg.ID), zap.Any("memos", command.Memo))
		if len(command.Memo) == 0 {
			return nil
		}

		err := wp.env.UpsertMemo(command.Memo)
		if err != nil {
			return errors.E(op, err)
		}

	default:
		return errors.E(op, errors.Str("undefined command"))
	}

	return nil
}

func (wp *Workflow) createLocalActivityCallback(id uint64) bindings.LocalActivityResultHandler {
	callback := func(lar *bindings.LocalActivityResultWrapper) {
		wp.log.Debug("executing local activity callback", zap.Uint64("ID", id))
		wp.canceller.Discard(id)

		if lar.Err != nil {
			wp.log.Debug("error", zap.Error(lar.Err), zap.Int32("attempt", lar.Attempt), zap.Duration("backoff", lar.Backoff))
			wp.mq.PushError(id, temporal.GetDefaultFailureConverter().ErrorToFailure(lar.Err))
			return
		}

		wp.log.Debug("pushing local activity response", zap.Uint64("ID", id))
		wp.mq.PushResponse(id, lar.Result)
	}

	return func(lar *bindings.LocalActivityResultWrapper) {
		// timer cancel callback can happen inside the loop
		if atomic.LoadUint32(&wp.inLoop) == 1 {
			wp.log.Debug("calling local activity callback IN LOOP", zap.Uint64("ID", id))
			callback(lar)
			return
		}

		wp.callbacks = append(wp.callbacks, func() error {
			wp.log.Debug("appending local activity callback", zap.Uint64("ID", id))
			callback(lar)
			return nil
		})
	}
}

func (wp *Workflow) createCallback(id uint64, t string) bindings.ResultHandler {
	callback := func(result *commonpb.Payloads, err error) {
		wp.log.Debug("executing callback", zap.Uint64("ID", id), zap.String("type", t))
		wp.canceller.Discard(id)

		if err != nil {
			wp.log.Debug("error", zap.Error(err), zap.String("type", t))
			wp.mq.PushError(id, temporal.GetDefaultFailureConverter().ErrorToFailure(err))
			return
		}

		wp.log.Debug("pushing response", zap.Uint64("ID", id), zap.String("type", t))
		// fetch original payload
		wp.mq.PushResponse(id, result)
	}

	return func(result *commonpb.Payloads, err error) {
		// timer cancel callback can happen inside the loop
		if atomic.LoadUint32(&wp.inLoop) == 1 {
			wp.log.Debug("calling callback IN LOOP", zap.Uint64("ID", id), zap.String("type", t))
			callback(result, err)
			return
		}

		wp.callbacks = append(wp.callbacks, func() error {
			wp.log.Debug("appending callback", zap.Uint64("ID", id), zap.String("type", t))
			callback(result, err)
			return nil
		})
	}
}

// callback to be called inside the queue processing, adds new messages at the end of the queue
func (wp *Workflow) createContinuableCallback(id uint64, t string) bindings.ResultHandler {
	callback := func(result *commonpb.Payloads, err error) {
		wp.log.Debug("executing continuable callback", zap.Uint64("ID", id), zap.String("type", t))
		wp.canceller.Discard(id)

		if err != nil {
			wp.mq.PushError(id, temporal.GetDefaultFailureConverter().ErrorToFailure(err))
			return
		}

		wp.mq.PushResponse(id, result)
		err = wp.flushQueue()
		if err != nil {
			panic(err)
		}
	}

	return func(result *commonpb.Payloads, err error) {
		callback(result, err)
	}
}

// Exchange messages between host and pool processes and add new commands to the queue.
func (wp *Workflow) flushQueue() error {
	const op = errors.Op("flush_queue")

	if len(wp.mq.Messages()) == 0 {
		return nil
	}

	if wp.mh != nil {
		wp.mh.Gauge(RrWorkflowsMetricName).Update(float64(wp.pool.QueueSize()))
		defer wp.mh.Gauge(RrWorkflowsMetricName).Update(float64(wp.pool.QueueSize()))
	}

	pl := wp.getPld()
	defer wp.putPld(pl)
	err := wp.codec.Encode(wp.getContext(), pl, wp.mq.Messages()...)
	if err != nil {
		return err
	}

	ch := make(chan struct{}, 1)
	result, err := wp.pool.Exec(context.Background(), pl, ch)
	if err != nil {
		return err
	}

	var r *payload.Payload
	select {
	case pld := <-result:
		if pld.Error() != nil {
			return errors.E(op, pld.Error())
		}
		// streaming is not supported
		if pld.Payload().Flags&frame.STREAM != 0 {
			ch <- struct{}{}
			return errors.E(op, errors.Str("streaming is not supported"))
		}

		// assign the payload
		r = pld.Payload()
	default:
		return errors.E(op, errors.Str("worker empty response"))
	}

	msgs := make([]*internal.Message, 0, 2)
	err = wp.codec.Decode(r, &msgs)
	if err != nil {
		return err
	}
	wp.mq.Flush()
	wp.pipeline = append(wp.pipeline, msgs...)

	return nil
}

// Run single command and return a single result.
func (wp *Workflow) runCommand(cmd any, payloads *commonpb.Payloads, header *commonpb.Header) (*internal.Message, error) {
	const op = errors.Op("workflow_process_runcommand")
	msg := &internal.Message{}
	wp.mq.AllocateMessage(cmd, payloads, header, msg)

	if wp.mh != nil {
		wp.mh.Gauge(RrMetricName).Update(float64(wp.pool.QueueSize()))
		defer wp.mh.Gauge(RrMetricName).Update(float64(wp.pool.QueueSize()))
	}

	pl := wp.getPld()
	err := wp.codec.Encode(wp.getContext(), pl, msg)
	if err != nil {
		wp.putPld(pl)
		return nil, err
	}

	// todo(rustatian): do we need a timeout here??
	ch := make(chan struct{}, 1)
	result, err := wp.pool.Exec(context.Background(), pl, ch)
	if err != nil {
		wp.putPld(pl)
		return nil, err
	}

	var r *payload.Payload
	select {
	case pld := <-result:
		if pld.Error() != nil {
			return nil, errors.E(op, pld.Error())
		}
		// streaming is not supported
		if pld.Payload().Flags&frame.STREAM != 0 {
			ch <- struct{}{}
			return nil, errors.E(op, errors.Str("streaming is not supported"))
		}

		// assign the payload
		r = pld.Payload()
	default:
		return nil, errors.E(op, errors.Str("worker empty response"))
	}

	msgs := make([]*internal.Message, 0, 2)
	err = wp.codec.Decode(r, &msgs)
	if err != nil {
		wp.putPld(pl)
		return nil, err
	}

	if len(msgs) != 1 {
		wp.putPld(pl)
		return nil, errors.E(op, errors.Str("unexpected pool response"))
	}

	wp.putPld(pl)
	return msgs[0], nil
}

func (wp *Workflow) getPld() *payload.Payload {
	return wp.pldPool.Get().(*payload.Payload)
}

func (wp *Workflow) putPld(pld *payload.Payload) {
	pld.Codec = 0
	pld.Context = nil
	pld.Body = nil
	wp.pldPool.Put(pld)
}
