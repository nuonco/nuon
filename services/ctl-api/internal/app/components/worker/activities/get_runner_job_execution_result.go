package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

type GetRunnerJobExecutionResultRequest struct {
	RunnerJobExecutionID string `validate:"required"`
}

// @temporal-gen activity
// @max-retries 1
func (a *Activities) GetRunnerJobExecutionResult(ctx context.Context, req *GetRunnerJobExecutionResultRequest) (*app.RunnerJobExecutionResult, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("runner_job_execution_id", req.RunnerJobExecutionID))
	l.Info("fetching runner job execution result")

	var result app.RunnerJobExecutionResult

	res := a.db.WithContext(ctx).
		Where("runner_job_execution_id = ?", req.RunnerJobExecutionID).
		Order("created_at DESC").
		First(&result)
	if res.Error != nil {
		l.Error("unable to get runner job execution result", zap.Error(res.Error))
		return nil, fmt.Errorf("unable to get runner job execution result: %w", res.Error)
	}

	l.Info("runner job execution result fetched")

	return &result, nil
}
