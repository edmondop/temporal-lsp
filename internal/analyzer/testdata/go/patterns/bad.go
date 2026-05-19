package patterns

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"
)

func BadWorkflow(ctx workflow.Context) error {
	// no-context-propagation: using context.Background() in a workflow
	actCtx := context.Background()
	_ = actCtx

	// activity-timeout-required: ExecuteActivity without timeout in options
	var result string
	workflow.ExecuteActivity(ctx, "MyActivity", "input").Get(ctx, &result)

	// unbounded-loop: for loop without ContinueAsNew
	for {
		workflow.Sleep(ctx, 0)
	}
}

func BadActivity(ctx context.Context) error {
	// no-naked-error: returning a bare error without temporal.NewApplicationError
	return fmt.Errorf("something failed")
}
