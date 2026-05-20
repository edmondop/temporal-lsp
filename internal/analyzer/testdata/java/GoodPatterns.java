package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import io.temporal.workflow.Workflow;
import io.temporal.activity.ActivityOptions;
import java.time.Duration;

@WorkflowInterface
public interface GoodPatternsWorkflow {
    @WorkflowMethod
    String run(String input);
}

public class GoodPatternsImpl implements GoodPatternsWorkflow {
    private final MyActivity activity = Workflow.newActivityStub(
        MyActivity.class,
        ActivityOptions.newBuilder()
            .withStartToCloseTimeout(Duration.ofSeconds(30))
            .build()
    );

    @WorkflowMethod
    public String run(String input) {
        activity.execute(input);
        while (true) {
            Workflow.sleep(Duration.ofSeconds(60));
            Workflow.continueAsNew(input);
        }
    }
}
