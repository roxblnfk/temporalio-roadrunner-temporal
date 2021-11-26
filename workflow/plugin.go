package workflow

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner/v2/events"
	"github.com/spiral/roadrunner/v2/state/process"
	roadrunner_temporal "github.com/temporalio/roadrunner-temporal"
	"github.com/temporalio/roadrunner-temporal/client"
)

const (
	// PluginName defines public service name.
	PluginName = "workflows"
)

// Plugin manages workflows and workers.
type Plugin struct {
	// embed
	sync.Mutex
	// plugins
	temporal client.Temporal

	events events.EventBus
	id     string

	server server.Server
	log    logger.Logger

	// graceful timeout for the worker
	gracePeriod time.Duration

	reset   chan struct{}
	stopCh  chan struct{}
	pool    pool
	closing int64
}

// Init workflow plugin.
func (p *Plugin) Init(temporal client.Temporal, server server.Server, log logger.Logger, cfg config.Configurer) error {
	const op = errors.Op("workflow_plugin_init")
	if !cfg.Has(roadrunner_temporal.RootPluginName) {
		return errors.E(op, errors.Disabled)
	}
	p.temporal = temporal
	p.server = server
	p.events, p.id = events.Bus()
	p.log = log
	p.reset = make(chan struct{}, 1)
	p.stopCh = make(chan struct{}, 1)

	// it can't be 0 (except set by user), because it would be set by the rr-binary (cli)
	p.gracePeriod = cfg.GetCommonConfig().GracefulTimeout

	return nil
}

// Serve starts workflow service.
func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	const op = errors.Op("workflow_plugin_serve")

	p.Lock()
	defer p.Unlock()

	var err error
	p.pool, err = p.startPool()
	if err != nil {
		errCh <- errors.E(op, err)
		return errCh
	}

	// start pool watcher
	go p.watch(errCh)

	return errCh
}

// Stop workflow service.
func (p *Plugin) Stop() error {
	atomic.StoreInt64(&p.closing, 1)
	const op = errors.Op("workflow_plugin_stop")

	workflowPool := p.getPool()
	if workflowPool != nil {
		err := workflowPool.Destroy(context.Background())
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	p.stopCh <- struct{}{}

	return nil
}

// Name of the service.
func (p *Plugin) Name() string {
	return PluginName
}

// Workers returns list of available workflow workers.
func (p *Plugin) Workers() []*process.State {
	p.Lock()
	defer p.Unlock()
	if p.pool == nil {
		return nil
	}

	workers := p.pool.Workers()
	states := make([]*process.State, 0, len(workers))

	for i := 0; i < len(workers); i++ {
		st, err := process.WorkerProcessState(workers[i])
		if err != nil {
			// log error and continue
			p.log.Error("worker process state error", "error", err)
			continue
		}

		states = append(states, st)
	}

	return states
}

func (p *Plugin) Available() {}

// WorkflowNames returns list of all available workflows.
func (p *Plugin) WorkflowNames() []string {
	return p.pool.WorkflowNames()
}

// Reset resets underlying workflow pool with new copy.
func (p *Plugin) Reset() error {
	p.reset <- struct{}{}
	return nil
}

// AddListener adds event listeners to the service.
func (p *Plugin) startPool() (pool, error) {
	const op = errors.Op("workflow_plugin_start_worker")
	np, err := newPool(
		p.temporal.GetCodec().WithLogger(p.log),
		p.server,
		p.gracePeriod,
		p.log,
	)
	if err != nil {
		return nil, errors.E(op, err)
	}

	err = np.Start(context.Background(), p.temporal)
	if err != nil {
		return nil, errors.E(op, err)
	}

	p.log.Debug("Started workflow processing", "workflows", np.WorkflowNames())

	return np, nil
}

func (p *Plugin) replacePool() error {
	p.Lock()
	defer p.Unlock()

	const op = errors.Op("workflow_plugin_replace_worker")
	if p.pool != nil {
		err := p.pool.Destroy(context.Background())
		p.pool = nil
		if err != nil {
			p.log.Error(
				"Unable to destroy expired workflow pool",
				"error",
				errors.E(op, err),
			)
			return errors.E(op, err)
		}
	}

	np, err := p.startPool()
	if err != nil {
		p.log.Error("Replace workflow pool failed", "error", err)
		return errors.E(op, err)
	}

	p.pool = np
	p.log.Debug("workflow pool successfully replaced")

	return nil
}

// getPool returns pool.
func (p *Plugin) getPool() pool {
	p.Lock()
	defer p.Unlock()

	return p.pool
}

// watch takes care about replacing pool
func (p *Plugin) watch(errCh chan error) {
	go func() {
		const op = errors.Op("workflow_plugin_watch")
		for range p.reset {
			if atomic.LoadInt64(&p.closing) == 1 {
				return
			}

			err := p.replacePool()
			if err == nil {
				continue
			}

			bkoff := backoff.NewExponentialBackOff()
			bkoff.InitialInterval = time.Second

			err = backoff.Retry(p.replacePool, bkoff)
			if err != nil {
				p.log.Error("failed to replace workflow pool", "error", errors.E(op, err))
				errCh <- err
				return
			}
		}
	}()

	go func() {
		eventsCh := make(chan events.Event, 10)
		err := p.events.SubscribeP(p.id, "*.EventWorkerWaitExit", eventsCh)
		if err != nil {
			errCh <- err
			return
		}

		err = p.events.SubscribeP(p.id, "*.EventWorkerError", eventsCh)
		if err != nil {
			errCh <- err
			return
		}

		for {
			select {
			case <-eventsCh:
				p.reset <- struct{}{}
			case <-p.stopCh:
				p.events.Unsubscribe(p.id)
				return
			}
		}
	}()
}
