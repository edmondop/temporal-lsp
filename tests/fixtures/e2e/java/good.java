package com.example;

import io.temporal.workflow.WorkflowMethod;
import io.temporal.workflow.Workflow;

public interface GreetingWorkflow {
    @WorkflowMethod
    String greet(String name) {
        Workflow.sleep(java.time.Duration.ofSeconds(1));
        return "Hello " + name;
    }
}
