package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import io.temporal.activity.ActivityInterface;
import io.temporal.activity.ActivityMethod;

@WorkflowInterface
public interface BadSignaturesWorkflow {
    @WorkflowMethod
    String run(String name, int age, boolean active);
}

@ActivityInterface
public interface BadSignaturesActivity {
    @ActivityMethod
    String process(String id, int count);
}
