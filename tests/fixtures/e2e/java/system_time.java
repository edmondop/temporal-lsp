package com.example;

import io.temporal.workflow.WorkflowMethod;

public interface TimerWorkflow {
    @WorkflowMethod
    long getTimestamp() {
        return System.currentTimeMillis();
    }
}
