package workflows

import (
	"go.temporal.io/sdk/workflow"
)

func PollingWorkflow(ctx workflow.Context) error {
	for {
		err := workflow.ExecuteActivity(ctx, "CheckStatus").Get(ctx, nil)
		if err != nil {
			return err
		}
	}
}
