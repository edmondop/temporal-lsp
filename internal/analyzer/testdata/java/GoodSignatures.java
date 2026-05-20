package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;

@WorkflowInterface
public interface GoodSignaturesWorkflow {
    @WorkflowMethod
    String run(WorkflowInput input);
}

@ActivityInterface
public interface GoodSignaturesActivity {
    @ActivityMethod
    String process(ActivityInput input);
}
