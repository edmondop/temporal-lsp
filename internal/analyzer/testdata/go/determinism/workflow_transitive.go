package determinism

import "go.temporal.io/sdk/workflow"

func TransitiveWorkflow(ctx workflow.Context) error {
	_ = getCurrentTime()
	return nil
}
