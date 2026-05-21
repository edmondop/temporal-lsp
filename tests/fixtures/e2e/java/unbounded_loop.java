package com.example;

import io.temporal.workflow.WorkflowMethod;

public interface PollingWorkflow {
    @WorkflowMethod
    void poll() {
        while (true) {
            checkStatus();
        }
    }
}
