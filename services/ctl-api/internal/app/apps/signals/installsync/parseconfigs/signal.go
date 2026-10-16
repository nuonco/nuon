package parseconfigs

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	activities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

const SignalType signal.SignalType = "install-sync-parse-configs"

type Signal struct {
	AppID                  string `json:"app_id" validate:"required"`
	AppInstallConfigSyncID string `json:"app_install_config_sync_id" validate:"required"`
	InstallsDirectory      string `json:"installs_directory,omitempty"`
	CommitSHA              string `json:"commit_sha,omitempty"`

	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithOnApprove = (*Signal)(nil)

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
	return nil
}

func (s *Signal) OnApprove(ctx workflow.Context) error {
	step, err := activities.AwaitPkgWorkflowsFlowGetFlowsStepByFlowStepID(ctx, s.StepID)
	if err != nil {
		return fmt.Errorf("unable to get step: %w", err)
	}

	rawProposals, ok := step.Status.Metadata["proposed_installs"]
	if !ok {
		return fmt.Errorf("proposed_installs not found in step metadata")
	}

	data, err := json.Marshal(rawProposals)
	if err != nil {
		return fmt.Errorf("unable to encode proposed installs: %w", err)
	}
	var proposals []app.ProposedInstall
	if err := json.Unmarshal(data, &proposals); err != nil {
		return fmt.Errorf("unable to decode proposed installs: %w", err)
	}

	if len(proposals) == 0 {
		return nil
	}

	_, err = branchactivities.AwaitCreateProposedInstalls(ctx, &branchactivities.CreateProposedInstallsInput{
		AppID: s.AppID, AppInstallConfigSyncID: s.AppInstallConfigSyncID,
		InstallConfigSyncID: "ics" + s.AppInstallConfigSyncID[3:],
		Installs:            proposals,
	})
	if err != nil {
		return fmt.Errorf("unable to create proposed installs: %w", err)
	}

	return nil
}
