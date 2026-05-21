package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import io.temporal.workflow.Workflow;

@WorkflowInterface
public interface GoodWorkflowInterface {
    @WorkflowMethod
    String run(WorkflowInput input);
}

public class GoodWorkflow implements GoodWorkflowInterface {
    @WorkflowMethod
    public String run(WorkflowInput input) {
        long now = Workflow.currentTimeMillis();
        Workflow.sleep(java.time.Duration.ofSeconds(1));
        int r = Workflow.newRandom().nextInt();
        return "done";
    }
}
