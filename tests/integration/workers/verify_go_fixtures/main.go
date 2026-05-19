// verify_go_fixtures connects to a Temporal server and registers the
// fixture workflows/activities to prove they are real Temporal code.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"verify_go_fixtures/determinism"
	"verify_go_fixtures/patterns"
	"verify_go_fixtures/signatures"
)

func logf(format string, args ...any) {
	fmt.Fprintf(os.Stdout, "[verify] "+format+"\n", args...)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: verify_go_fixtures <temporal_address>")
		os.Exit(1)
	}
	address := os.Args[1]
	logf("Starting verification against %s (pid=%d)", address, os.Getpid())

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Wait for TCP connectivity first (Temporal might still be starting)
	logf("Waiting for TCP connectivity to %s...", address)
	for i := 0; i < 60; i++ {
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			conn.Close()
			logf("TCP connection successful after %d attempts", i+1)
			goto connected
		}
		if i%5 == 0 {
			logf("  TCP attempt %d: %v", i+1, err)
		}
		time.Sleep(time.Second)
	}
	logf("FATAL: TCP connection to %s failed after 60 attempts", address)
	os.Exit(1)

connected:
	// Small delay for Temporal to finish initializing after port opens
	time.Sleep(2 * time.Second)

	logf("Creating Temporal client...")
	c, err := client.Dial(client.Options{HostPort: address})
	if err != nil {
		logf("FATAL: client.Dial failed: %v", err)
		os.Exit(1)
	}
	logf("Temporal client created")
	defer c.Close()

	// Register determinism fixtures
	logf("Registering determinism fixtures...")
	w1 := worker.New(c, "verify-determinism", worker.Options{})
	w1.RegisterWorkflow(determinism.MyWorkflow)
	w1.RegisterWorkflow(determinism.TransitiveWorkflow)
	logf("OK: determinism fixtures registered")

	// Register signature fixtures
	logf("Registering signature fixtures...")
	w2 := worker.New(c, "verify-signatures", worker.Options{})
	w2.RegisterWorkflow(signatures.BadWorkflow)
	w2.RegisterActivity(signatures.BadActivity)
	w2.RegisterWorkflow(signatures.GoodWorkflow)
	w2.RegisterActivity(signatures.GoodActivity)
	logf("OK: signature fixtures registered")

	// Register pattern fixtures
	logf("Registering pattern fixtures...")
	w3 := worker.New(c, "verify-patterns", worker.Options{})
	w3.RegisterWorkflow(patterns.BadWorkflow)
	w3.RegisterActivity(patterns.BadActivity)
	w3.RegisterWorkflow(patterns.GoodWorkflow)
	w3.RegisterActivity(patterns.GoodActivity)
	logf("OK: pattern fixtures registered")

	// Start workers briefly to prove they can handle tasks
	logf("Starting workers...")
	go func() { _ = w1.Run(worker.InterruptCh()) }()
	go func() { _ = w2.Run(worker.InterruptCh()) }()
	go func() { _ = w3.Run(worker.InterruptCh()) }()

	// Give workers a moment to start
	time.Sleep(2 * time.Second)
	logf("Workers started, executing GoodWorkflow...")

	// Execute GoodWorkflow from signatures to prove it works end-to-end
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "verify-good-signatures",
		TaskQueue: "verify-signatures",
	}, signatures.GoodWorkflow, signatures.WorkflowInput{Name: "test", Age: 30})
	if err != nil {
		logf("FATAL: Failed to execute GoodWorkflow: %v", err)
		os.Exit(1)
	}
	logf("Workflow started (ID=%s, RunID=%s), waiting for result...", run.GetID(), run.GetRunID())

	var result signatures.WorkflowResult
	if err := run.Get(ctx, &result); err != nil {
		logf("FATAL: GoodWorkflow execution failed: %v", err)
		os.Exit(1)
	}
	logf("OK: GoodWorkflow executed, result=%+v", result)

	fmt.Println("\nAll Go fixtures verified as real Temporal workflows.")
	os.Exit(0)
}
