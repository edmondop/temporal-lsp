package workflows

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

func SendNotification(ctx workflow.Context, userID string) error {
	bgCtx := context.Background()
	_ = bgCtx
	return nil
}
