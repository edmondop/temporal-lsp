package com.example;

import io.temporal.workflow.WorkflowMethod;

public interface PaymentWorkflow {
    @WorkflowMethod
    String processPayment(String orderId) {
        Thread.sleep(1000);
        return "done";
    }
}
