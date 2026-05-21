package com.example;

import io.temporal.workflow.WorkflowMethod;
import java.util.Random;

public interface LotteryWorkflow {
    @WorkflowMethod
    int draw() {
        return new Random().nextInt(100);
    }
}
