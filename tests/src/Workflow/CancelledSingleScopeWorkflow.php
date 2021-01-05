<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Exception\CancellationException;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

class CancelledSingleScopeWorkflow
{
    private array $status = [];

    #[Workflow\QueryMethod(name: 'getStatus')]
    public function getStatus(): array
    {
        return $this->status;
    }

    #[WorkflowMethod(name: 'CancelledSingleScopeWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        $this->status[] = 'start';
        try {
            yield Workflow::newCancellationScope(
                function () use ($simple) {
                    try {
                        $this->status[] = 'in scope';
                        yield $simple->slow('1');
                    } catch (CancellationException $e) {
                        error_log("HANDLED IN SCOPE");
                        // after process is complete, do not use for business logic
                        $this->status[] = 'captured in scope';
                    }
                }
            )->onCancel(
                function () use (&$cancelled) {
                    $this->status[] = 'on cancel';
                }
            );
        } catch (CancellationException $e) {
            $this->status[] = 'captured in process';
        }

        return 'OK';
    }
}