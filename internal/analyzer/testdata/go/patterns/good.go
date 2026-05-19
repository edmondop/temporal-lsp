package patterns

import (
	"context"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func GoodWorkflow(ctx workflow.Context) error {
	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	actCtx := workflow.WithActivityOptions(ctx, actOpts)

	var result string
	workflow.ExecuteActivity(actCtx, "MyActivity", "input").Get(ctx, &result)

	// Bounded loop with ContinueAsNew
	for i := 0; i < 100; i++ {
		workflow.Sleep(ctx, 0)
	}
	return workflow.NewContinueAsNewError(ctx, GoodWorkflow)
}

func GoodActivity(ctx context.Context) error {
	return temporal.NewApplicationError("something failed", "MY_ERROR")
}
