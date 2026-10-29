package pininstallgroup

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-pin-install-group"

// Signal pins an app config version onto every install in a group without
// deploying it: it records an InstallAppConfigVersion row and enqueues
// update-app-config so the install's app_config_id moves forward, but it
// never creates an install workflow and never cancels anything in flight.
type Signal struct {
	InstallGroupID string `json:"install_group_id" validate:"required"`
	AppBranchID    string `json:"app_branch_id" validate:"required"`
	RunID          string `json:"run_id" validate:"required"`

	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithEmptyGroupCheck = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

// IsEmptyInstallGroup reports whether this group resolves to zero installs;
// empty groups are auto-skipped (see checks/emptygroup), matching
// planinstallgroup and updateinstallgroup.
func (s *Signal) IsEmptyInstallGroup(ctx workflow.Context) (bool, error) {
	installIDs, _, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return false, err
	}
	return len(installIDs) == 0, nil
}

func (s *Signal) Validate(ctx workflow.Context) error {
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	if _, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID); err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	if _, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID); err != nil {
		return errors.Wrap(err, "install group not found")
	}

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return errors.Wrap(err, "app branch run not found")
	}

	if run.AppConfigID == "" && !(run.NoConfigChanges && !run.Force) {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	return nil
}

// Cancel does not touch any in-flight work: pinning never creates install
// workflows, so there is nothing to cancel.
func (s *Signal) Cancel(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("pin install group cancelled, nothing to clean up",
		"install_group_id", s.InstallGroupID,
	)
	return nil
}
