package com.example.workflows;

import io.temporal.workflow.WorkflowInterface;
import io.temporal.workflow.WorkflowMethod;
import java.time.Instant;
import java.util.Date;
import java.util.Random;

@WorkflowInterface
public interface BadWorkflowInterface {
    @WorkflowMethod
    String run(String input);
}

public class BadWorkflow implements BadWorkflowInterface {
    @WorkflowMethod
    public String run(String input) {
        long now = System.currentTimeMillis();
        Thread.sleep(1000);
        double r = Math.random();
        new Random().nextInt();
        new Thread(() -> {}).start();
        new ReentrantLock();
        return "done";
    }
}
