package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.uber.org/zap"
)

type GetRunnerJobExecutionByJobIDRequest struct {
	JobID string `validate:"required"`
}

// @temporal-gen activity
// @max-retries 1
func (a *Activities) GetRunnerJobExecutionByJobID(ctx context.Context, req *GetRunnerJobExecutionByJobIDRequest) (*app.RunnerJobExecution, error) {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(zap.String("runner_job_id", req.JobID))
	l.Info("fetching runner job execution")

	var execution app.RunnerJobExecution
	res := a.db.WithContext(ctx).
		Where("runner_job_id = ?", req.JobID).
		Order("created_at DESC").
		First(&execution)
	if res.Error != nil {
		l.Error("unable to get runner job execution", zap.Error(res.Error))
		return nil, fmt.Errorf("unable to get runner job execution: %w", res.Error)
	}

	l.Info("runner job execution fetched")
	return &execution, nil
}
