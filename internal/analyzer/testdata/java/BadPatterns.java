package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import io.temporal.workflow.Workflow;
import io.temporal.activity.ActivityOptions;

@WorkflowInterface
public interface BadPatternsWorkflow {
    @WorkflowMethod
    String run(String input);
}

public class BadPatternsImpl implements BadPatternsWorkflow {
    private final MyActivity activity = Workflow.newActivityStub(
        MyActivity.class,
        ActivityOptions.newBuilder().build()
    );

    @WorkflowMethod
    public String run(String input) {
        activity.execute(input);
        while (true) {
            Workflow.sleep(java.time.Duration.ofSeconds(60));
        }
    }
}
