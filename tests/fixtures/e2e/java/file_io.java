package com.example;

import io.temporal.workflow.WorkflowMethod;
import java.io.File;

public interface BackupWorkflow {
    @WorkflowMethod
    boolean check() {
        return new File("/tmp/data.txt").exists();
    }
}
