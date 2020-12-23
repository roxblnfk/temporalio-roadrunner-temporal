<?php

namespace Temporal\Tests\Workflow;

use React\Promise\Deferred;
use Temporal\Client\Workflow;
use Temporal\Client\Workflow\WorkflowMethod;

class RuntimeSignalWorkflow
{
    #[WorkflowMethod(name: 'RuntimeSignalWorkflow')]
    public function handler()
    {
        $wait = new Deferred();

        $counter = 0;

        Workflow::registerSignal('add', function ($value) use (&$counter, $wait) {
            if (is_array($value)) {
                // todo: fix it
                $value = $value[0];
            }

            $counter += $value;
            $wait->resolve($value);
        });

        yield $wait;

        return $counter;
    }
}