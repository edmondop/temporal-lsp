package workflows

import (
	"go.temporal.io/sdk/workflow"
)

func ProcessOrder(ctx workflow.Context, orderID string, quantity int, price float64) (string, int, error) {
	return "done", 0, nil
}
