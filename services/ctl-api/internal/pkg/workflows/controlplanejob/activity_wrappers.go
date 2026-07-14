package controlplanejob

import "go.temporal.io/sdk/workflow"

func AwaitEnsureExecution(ctx workflow.Context, req *EnsureExecutionRequest) (*EnsureExecutionResponse, error) {
	var out *EnsureExecutionResponse
	err := workflow.ExecuteActivity(ctx, (*Activities).EnsureExecution, req).Get(ctx, &out)
	return out, err
}

func AwaitRunJob(ctx workflow.Context, req *RunJobRequest) error {
	return workflow.ExecuteActivity(ctx, (*Activities).RunJob, req).Get(ctx, nil)
}

func AwaitFinalize(ctx workflow.Context, req *FinalizeRequest) (*FinalizeResponse, error) {
	var out *FinalizeResponse
	err := workflow.ExecuteActivity(ctx, (*Activities).Finalize, req).Get(ctx, &out)
	return out, err
}
