package signatures

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

// Multiple primitive params — triggers primitive-params and single-payload
func BadWorkflow(ctx workflow.Context, name string, age int) (string, error) {
	return "", nil
}

// Too many return values — triggers single-return
func BadActivity(ctx context.Context, id string) (string, int, error) {
	return "", 0, nil
}
