package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;

@WorkflowInterface
public interface BatchWorkflow {
    @WorkflowMethod
    String run(String batchId);
}

public class BatchWorkflowImpl implements BatchWorkflow {
    @WorkflowMethod
    public String run(String batchId) {
        new Thread(() -> processBatch(batchId)).start();
        return "started: " + batchId;
    }
}
