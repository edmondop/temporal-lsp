// verify_go_fixtures connects to a Temporal server and registers the
// fixture workflows/activities to prove they are real Temporal code.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"verify_go_fixtures/determinism"
	"verify_go_fixtures/patterns"
	"verify_go_fixtures/signatures"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: verify_go_fixtures <temporal_address>")
	}
	address := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Retry connection for up to 20 seconds (Temporal might still be starting)
	var c client.Client
	var err error
	for i := 0; i < 20; i++ {
		c, err = client.Dial(client.Options{HostPort: address})
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatalf("Failed to connect to Temporal at %s: %v", address, err)
	}
	defer c.Close()

	// Register determinism fixtures
	w1 := worker.New(c, "verify-determinism", worker.Options{})
	w1.RegisterWorkflow(determinism.MyWorkflow)
	w1.RegisterWorkflow(determinism.TransitiveWorkflow)
	fmt.Println("OK: determinism fixtures registered")

	// Register signature fixtures
	w2 := worker.New(c, "verify-signatures", worker.Options{})
	w2.RegisterWorkflow(signatures.BadWorkflow)
	w2.RegisterActivity(signatures.BadActivity)
	w2.RegisterWorkflow(signatures.GoodWorkflow)
	w2.RegisterActivity(signatures.GoodActivity)
	fmt.Println("OK: signature fixtures registered")

	// Register pattern fixtures
	w3 := worker.New(c, "verify-patterns", worker.Options{})
	w3.RegisterWorkflow(patterns.BadWorkflow)
	w3.RegisterActivity(patterns.BadActivity)
	w3.RegisterWorkflow(patterns.GoodWorkflow)
	w3.RegisterActivity(patterns.GoodActivity)
	fmt.Println("OK: pattern fixtures registered")

	// Start workers briefly to prove they can handle tasks
	go func() { _ = w1.Run(worker.InterruptCh()) }()
	go func() { _ = w2.Run(worker.InterruptCh()) }()
	go func() { _ = w3.Run(worker.InterruptCh()) }()

	// Give workers a moment to start
	time.Sleep(2 * time.Second)

	// Execute GoodWorkflow from signatures to prove it works end-to-end
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "verify-good-signatures",
		TaskQueue: "verify-signatures",
	}, signatures.GoodWorkflow, signatures.WorkflowInput{Name: "test", Age: 30})
	if err != nil {
		log.Fatalf("Failed to execute GoodWorkflow: %v", err)
	}

	var result signatures.WorkflowResult
	if err := run.Get(ctx, &result); err != nil {
		log.Fatalf("GoodWorkflow execution failed: %v", err)
	}
	fmt.Printf("OK: GoodWorkflow executed, result=%+v\n", result)

	fmt.Println("\nAll Go fixtures verified as real Temporal workflows.")
	os.Exit(0)
}
