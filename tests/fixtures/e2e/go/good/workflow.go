package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

type OrderInput struct {
	OrderID string
	Amount  float64
}

type OrderResult struct {
	Status    string
	Timestamp string
}

func ProcessOrder(ctx workflow.Context, input OrderInput) (OrderResult, error) {
	opts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, opts)

	var result OrderResult
	err := workflow.ExecuteActivity(ctx, "ProcessPayment", input).Get(ctx, &result)
	return result, err
}
