package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;

@WorkflowInterface
public interface RetryWorkflow {
    @WorkflowMethod
    String run(String taskId);
}

public class RetryWorkflowImpl implements RetryWorkflow {
    @WorkflowMethod
    public String run(String taskId) {
        Thread.sleep(5000);
        return "retried: " + taskId;
    }
}
