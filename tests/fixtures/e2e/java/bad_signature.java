package com.example;

import io.temporal.workflow.WorkflowMethod;
import io.temporal.activity.ActivityMethod;

public interface OrderWorkflow {
    @WorkflowMethod
    String process(String orderId, int quantity, double price);
}

public interface OrderActivities {
    @ActivityMethod
    boolean sendEmail(String to, String subject, String body);
}
