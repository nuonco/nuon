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
)

type RecordPolicyEvaluationCompositeErrorRequest struct {
	WorkflowStepID string `validate:"required"`
	StepTargetID   string `validate:"required"`
	StepTargetType string `validate:"required"`
	Stage          policyerrors.EvaluationFailureStage
}

// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) RecordPolicyEvaluationCompositeError(ctx context.Context, req RecordPolicyEvaluationCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("workflow_step_id", req.WorkflowStepID),
		zap.String("step_target_id", req.StepTargetID),
		zap.String("step_target_type", req.StepTargetType),
		zap.String("stage", string(req.Stage)),
	)

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
