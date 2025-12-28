package examples

import (
	"go.temporal.io/sdk/workflow"
)

// SimpleWorkflow is a basic workflow
// @temporal-gen-v2 workflow
func SimpleWorkflow(ctx workflow.Context, input string) (string, error) {
	return "done", nil
}

// ComplexWorkflow demonstrates all available workflow options
// @temporal-gen-v2 workflow
// @execution-timeout 24h
// @task-timeout 10m
// @task-queue my-queue
// @wait-for-cancellation true
func ComplexWorkflow(ctx workflow.Context, input int) (int, error) {
	return input, nil
}

// WorkflowWithIDTemplate demonstrates custom ID templating
// @temporal-gen-v2 workflow
// @id-template workflow-{{.ID}}
func WorkflowWithIDTemplate(ctx workflow.Context, input IDInput) error {
	return nil
}

type IDInput struct {
	ID string
}

// WorkflowWithIDCallback demonstrates dynamic ID generation
// @temporal-gen-v2 workflow
// @id-generator GenerateWorkflowID
func WorkflowWithIDCallback(ctx workflow.Context, input string) error {
	return nil
}

func GenerateWorkflowID(input string) string {
	return "workflow-" + input
}

// WorkflowWithOptionsCallback demonstrates custom options callback
// @temporal-gen-v2 workflow
// @options-callback GetWorkflowOptions
func WorkflowWithOptionsCallback(ctx workflow.Context, input string) error {
	return nil
}

func GetWorkflowOptions(input string) *workflow.ChildWorkflowOptions {
	return &workflow.ChildWorkflowOptions{}
}
