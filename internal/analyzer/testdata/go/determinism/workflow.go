package determinism

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func MyWorkflow(ctx workflow.Context) error {
	_ = time.Now()
	return nil
}
