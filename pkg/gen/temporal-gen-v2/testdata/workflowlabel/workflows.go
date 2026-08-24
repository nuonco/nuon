// Package workflowlabel is a fixture for TestGenerateLabelOnWorkflowFails:
// labels apply to activities only, so @label on a workflow must be rejected
// rather than silently ignored.
package workflowlabel

import "go.temporal.io/sdk/workflow"

// @temporal-gen-v2 workflow
// @label tier critical
func LabeledWorkflow(ctx workflow.Context, input string) error {
	return nil
}
