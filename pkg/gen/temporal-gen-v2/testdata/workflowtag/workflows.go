// Package workflowtag is a fixture for TestGenerateTagOnWorkflowFails: tags
// apply to activities only, so @tag on a workflow must be rejected rather than
// silently ignored.
package workflowtag

import "go.temporal.io/sdk/workflow"

// @temporal-gen-v2 workflow
// @tag critical
func TaggedWorkflow(ctx workflow.Context, input string) error {
	return nil
}
