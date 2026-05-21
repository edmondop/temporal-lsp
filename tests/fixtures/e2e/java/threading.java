package com.example;

import io.temporal.workflow.WorkflowMethod;

public interface ParallelWorkflow {
    @WorkflowMethod
    void process() {
        new Thread(() -> doWork()).start();
    }
}
