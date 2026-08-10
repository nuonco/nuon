package postdeployrunbooks

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-post-deploy-runbooks"

type Signal struct {
	InstallGroupID string `json:"install_group_id" validate:"required"`
	AppBranchID    string `json:"app_branch_id" validate:"required"`
	RunID          string `json:"run_id" validate:"required"`

	// AppBranchConfigID pins the config this step was generated from. Resolving
	// the branch's latest config instead would let a mid-run sync or a saved
	// deployment plan change which runbooks execute.
	AppBranchConfigID string `json:"app_branch_config_id" validate:"required"`

	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`

	childWorkflowIDs []string
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	if _, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID); err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	if _, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID); err != nil {
		return errors.Wrap(err, "app branch run not found")
	}

	return nil
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("cancelling post-deploy runbooks",
		"install_group_id", s.InstallGroupID,
		"child_workflow_count", len(s.childWorkflowIDs),
	)

	for _, wfID := range s.childWorkflowIDs {
		if err := activities.AwaitCancelInstallWorkflow(ctx, &activities.CancelInstallWorkflowInput{
			WorkflowID: wfID,
		}); err != nil {
			logger.Warn("failed to cancel runbook run workflow",
				"workflow_id", wfID,
				"error", err,
			)
		}
	}

	return nil
}
