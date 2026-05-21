package com.example;

import io.temporal.workflow.WorkflowMethod;
import java.util.concurrent.locks.ReentrantLock;

public interface SyncWorkflow {
    @WorkflowMethod
    void process() {
        new ReentrantLock().lock();
    }
}
