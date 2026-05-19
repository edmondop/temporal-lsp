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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: verify_go_fixtures <temporal_address>")
		os.Exit(1)
	}
	address := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Wait for TCP connectivity (Temporal starts in background)
	for i := 0; i < 30; i++ {
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			conn.Close()
			break
		}
		if i == 29 {
			fmt.Printf("FATAL: cannot connect to %s\n", address)
			os.Exit(1)
		}
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)

	c, err := client.Dial(client.Options{HostPort: address})
	if err != nil {
		fmt.Printf("FATAL: client.Dial: %v\n", err)
		os.Exit(1)
	}
	defer c.Close()

	w1 := worker.New(c, "verify-determinism", worker.Options{})
	w1.RegisterWorkflow(determinism.MyWorkflow)
	w1.RegisterWorkflow(determinism.TransitiveWorkflow)
	fmt.Println("OK: determinism fixtures registered")

	// Register signature fixtures (BadActivity has invalid 3-value return — skip it)
	w2 := worker.New(c, "verify-signatures", worker.Options{})
	w2.RegisterWorkflow(signatures.BadWorkflow)
	w2.RegisterWorkflow(signatures.GoodWorkflow)
	w2.RegisterActivity(signatures.GoodActivity)
	fmt.Println("OK: signature fixtures registered")

	w3 := worker.New(c, "verify-patterns", worker.Options{})
	w3.RegisterWorkflow(patterns.BadWorkflow)
	w3.RegisterActivity(patterns.BadActivity)
	w3.RegisterWorkflow(patterns.GoodWorkflow)
	w3.RegisterActivity(patterns.GoodActivity)
	fmt.Println("OK: pattern fixtures registered")

	// Start workers and execute a workflow end-to-end
	go func() { _ = w1.Run(worker.InterruptCh()) }()
	go func() { _ = w2.Run(worker.InterruptCh()) }()
	go func() { _ = w3.Run(worker.InterruptCh()) }()
	time.Sleep(2 * time.Second)

	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        "verify-good-signatures",
		TaskQueue: "verify-signatures",
	}, signatures.GoodWorkflow, signatures.WorkflowInput{Name: "test", Age: 30})
	if err != nil {
		fmt.Printf("FATAL: ExecuteWorkflow: %v\n", err)
		os.Exit(1)
	}

	var result signatures.WorkflowResult
	if err := run.Get(ctx, &result); err != nil {
		fmt.Printf("FATAL: workflow failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("OK: GoodWorkflow executed, result=%+v\n", result)

	fmt.Println("\nAll Go fixtures verified as real Temporal workflows.")
	os.Exit(0)
}
