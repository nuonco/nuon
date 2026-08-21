package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/policy_reports/policyerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type RecordPolicyEvaluationCompositeErrorRequest struct {
	WorkflowStepID string                              `json:"workflow_step_id" temporaljson:"workflow_step_id,omitempty" validate:"required_without=RunnerJobID"`
	StepTargetID   string                              `json:"step_target_id" temporaljson:"step_target_id,omitempty" validate:"required_without=RunnerJobID"`
	StepTargetType string                              `json:"step_target_type" temporaljson:"step_target_type,omitempty" validate:"required_without=RunnerJobID"`
	RunnerJobID    string                              `json:"runner_job_id" temporaljson:"runner_job_id,omitempty"`
	Stage          policyerrors.EvaluationFailureStage `json:"stage" temporaljson:"stage,omitempty"`
}

// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) RecordPolicyEvaluationCompositeError(ctx context.Context, req RecordPolicyEvaluationCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("workflow_step_id", req.WorkflowStepID),
		zap.String("step_target_id", req.StepTargetID),
		zap.String("step_target_type", req.StepTargetType),
		zap.String("runner_job_id", req.RunnerJobID),
		zap.String("stage", string(req.Stage)),
	)

	if req.RunnerJobID != "" {
		return a.recordRunnerJobPolicyEvaluationCompositeError(ctx, l, req)
	}
	if req.WorkflowStepID == "" || req.StepTargetID == "" || req.StepTargetType == "" {
		return fmt.Errorf("workflow step id, step target id, and step target type are required")
	}

	data, err := compositeerrors.New(
		&policyerrors.EvaluationFailedError{Stage: req.Stage},
		compositeerrors.WithSource("workflow_steps", req.WorkflowStepID),
	)
	if err != nil {
		return fmt.Errorf("unable to build policy evaluation composite error: %w", err)
	}

	var res *gorm.DB
	switch app.WorkflowStepTargetType(req.StepTargetType) {
	case app.WorkflowStepTargetTypeInstallDeploy, app.WorkflowStepTargetTypeInstallDeploys:
		res = a.db.WithContext(ctx).
			Model(&app.InstallDeploy{ID: req.StepTargetID}).
			Select("composite_error").
			Updates(app.InstallDeploy{CompositeError: data})
	case app.WorkflowStepTargetTypeInstallSandboxRun, app.WorkflowStepTargetTypeInstallSandboxRuns:
		res = a.db.WithContext(ctx).
			Model(&app.InstallSandboxRun{ID: req.StepTargetID}).
			Select("composite_error").
			Updates(app.InstallSandboxRun{CompositeError: data})
	default:
		return fmt.Errorf("unsupported policy evaluation target type %q", req.StepTargetType)
	}
	if res.Error != nil {
		return fmt.Errorf("unable to record policy evaluation composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no policy evaluation target found for id %s: %w", req.StepTargetID, gorm.ErrRecordNotFound)
	}

	l.Info("recorded policy evaluation composite error")
	return nil
}

func (a *Activities) recordRunnerJobPolicyEvaluationCompositeError(ctx context.Context, l *zap.Logger, req RecordPolicyEvaluationCompositeErrorRequest) error {
	data, err := compositeerrors.New(
		&policyerrors.EvaluationFailedError{Stage: req.Stage},
		compositeerrors.WithSource("runner_jobs", req.RunnerJobID),
	)
	if err != nil {
		return fmt.Errorf("unable to build policy evaluation composite error: %w", err)
	}

	res := a.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Model(&app.RunnerJob{ID: req.RunnerJobID}).
		Select("composite_error").
		Updates(app.RunnerJob{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to record policy evaluation composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no runner job found for id %s: %w", req.RunnerJobID, gorm.ErrRecordNotFound)
	}

	l.Info("recorded policy evaluation composite error")
	return nil
}
