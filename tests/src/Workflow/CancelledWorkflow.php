<?php

namespace Temporal\Tests\Workflow;

use Temporal\Activity\ActivityOptions;
use Temporal\Exception\CancellationException;
use Temporal\Tests\Activity\SimpleActivity;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;

class CancelledWorkflow
{
    #[WorkflowMethod(name: 'CancelledWorkflow')]
    public function handler()
    {
        $simple = Workflow::newActivityStub(
            SimpleActivity::class,
            ActivityOptions::new()->withStartToCloseTimeout(5)
        );

        try {
            $slow = $simple->slow('DOING SLOW ACTIVITY');
            return yield $slow;
        } catch (CancellationException $e) {
            $first = yield $slow;
            if ($first !== 'doing slow activity') {
                return 'failed to complete';
            }

            $second = yield $simple->echo('rollback');
            if ($second !== 'ROLLBACK') {
                return 'failed to compensate ' . $second;
            }

            return 'CANCELLED';
        }
    }
}