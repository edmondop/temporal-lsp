package workflows

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func ProcessOrder(ctx workflow.Context, orderID string) (string, error) {
	now := time.Now()
	return "processed " + orderID + " at " + now.String(), nil
}
