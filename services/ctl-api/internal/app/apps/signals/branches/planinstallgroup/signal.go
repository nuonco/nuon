package planinstallgroup

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-plan-install-group"

type Signal struct {
	InstallGroupID string `json:"install_group_id"`
	AppBranchID    string `json:"app_branch_id" validate:"required"`
	RunID          string `json:"run_id" validate:"required"`

	PreviewInstallID   string `json:"preview_install_id,omitempty"`
	SyntheticGroupName string `json:"synthetic_group_name,omitempty"`

	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithEmptyGroupCheck = (*Signal)(nil)
var _ signal.SignalWithAutoApproveOnPoliciesPassing = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

// IsEmptyInstallGroup reports whether this group resolves to zero installs;
// empty groups are auto-skipped (see checks/emptygroup).
func (s *Signal) IsEmptyInstallGroup(ctx workflow.Context) (bool, error) {
	installIDs, _, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return false, err
	}
	return len(installIDs) == 0, nil
}

// AutoApproveOnPoliciesPassing reports whether the group opted into approving
// its own plan. Synthetic preview groups have no install group row to configure,
// so they always require a response.
func (s *Signal) AutoApproveOnPoliciesPassing(ctx workflow.Context) bool {
	if s.InstallGroupID == "" {
		return false
	}

	group, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil || group == nil {
		return false
	}
	return group.GetAutoApproveOnPoliciesPassing()
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}
	if s.InstallGroupID == "" && s.PreviewInstallID == "" {
		return fmt.Errorf("install_group_id or preview_install_id is required")
	}

	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	if s.InstallGroupID != "" {
		_, err = activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
		if err != nil {
			return errors.Wrap(err, "install group not found")
		}
	}

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return errors.Wrap(err, "app branch run not found")
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	return nil
}
