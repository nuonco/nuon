package examples

import (
	"context"
	"go.temporal.io/sdk/workflow"
)

type MethodActivities struct{}

// @temporal-gen-v2 activity
// @options-callback "GetActivityOptions"
func (a *MethodActivities) MyActivity(ctx context.Context, input string) (string, error) {
	return "result", nil
}

func GetActivityOptions(input string) *workflow.ActivityOptions {
	return &workflow.ActivityOptions{}
}
