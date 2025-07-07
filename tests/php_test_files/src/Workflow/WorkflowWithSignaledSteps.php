<?php


namespace Temporal\Tests\Workflow;

use React\Promise\Deferred;
use React\Promise\PromiseInterface;
use Temporal\Activity\ActivityOptions;
use Temporal\Workflow;
use Temporal\Workflow\WorkflowMethod;
use Temporal\Tests\Activity\SimpleActivity;

#[Workflow\WorkflowInterface]
class WorkflowWithSignaledSteps
{
    #[WorkflowMethod(name: 'WorkflowWithSignaledSteps')]
    public function handler()
    {
        $value = 0;
        Workflow::registerQuery('value', function () use (&$value) {
            return $value;
        });

        yield $this->promiseSignal('begin');
        $value++;

        yield $this->promiseSignal('next1');
        $value++;

        yield $this->promiseSignal('next2');
        $value++;

        return $value;
    }

    // is this correct?
    private function promiseSignal(string $name): PromiseInterface
    {
        $signal = new Deferred();
        Workflow::registerSignal($name, function ($value) use ($signal) {
            $signal->resolve($value);
        });

        return $signal->promise();
    }
}
