package activities

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/config"
)

var cloneRepoLocalActivityOpts = workflow.LocalActivityOptions{
	ScheduleToCloseTimeout: 5 * time.Minute,
	RetryPolicy: &temporal.RetryPolicy{
		MaximumAttempts: 1,
	},
}

var fetchConfigLocalActivityOpts = workflow.LocalActivityOptions{
	ScheduleToCloseTimeout: 5 * time.Minute,
	RetryPolicy: &temporal.RetryPolicy{
		MaximumAttempts: 1,
	},
}

func LocalAwaitCloneRepo(ctx workflow.Context, input CloneRepoRequest) (*CloneRepoResult, error) {
	var result *CloneRepoResult
	localCtx := workflow.WithLocalActivityOptions(ctx, cloneRepoLocalActivityOpts)
	err := workflow.ExecuteLocalActivity(localCtx,
		(*Activities).CloneRepo, input).Get(ctx, &result)
	return result, err
}

func LocalAwaitFetchIntermediateConfig(ctx workflow.Context, input FetchIntermediateConfigRequest) (*config.AppConfig, error) {
	var result *config.AppConfig
	localCtx := workflow.WithLocalActivityOptions(ctx, fetchConfigLocalActivityOpts)
	err := workflow.ExecuteLocalActivity(localCtx,
		(*Activities).FetchIntermediateConfig, input).Get(ctx, &result)
	return result, err
}
