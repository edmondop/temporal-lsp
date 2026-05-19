package signatures

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

type WorkflowInput struct {
	Name string
	Age  int
}

type WorkflowResult struct {
	Message string
}

type ActivityInput struct {
	ID string
}

type ActivityResult struct {
	Data string
}

// Struct params — no violations
func GoodWorkflow(ctx workflow.Context, input WorkflowInput) (WorkflowResult, error) {
	return WorkflowResult{}, nil
}

// Struct params — no violations
func GoodActivity(ctx context.Context, input ActivityInput) (ActivityResult, error) {
	return ActivityResult{}, nil
}

// Single primitive param is acceptable
func SimpleWorkflow(ctx workflow.Context, id string) error {
	return nil
}
