package syncinstalls

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "sync-installs"

type Signal struct {
	AppID   string `json:"app_id" validate:"required"`
	AppName string `json:"app_name,omitempty"`

	AppInstallConfigSyncID string `json:"app_install_config_sync_id,omitempty"`

	AppBranchID       string `json:"app_branch_id,omitempty"`
	AppBranchConfigID string `json:"app_branch_config_id,omitempty"`
	AppBranchRunID    string `json:"app_branch_run_id,omitempty"`

	CommitSHA   string `json:"commit_sha,omitempty"`
	TriggeredBy string `json:"triggered_by"`

	FallbackCreatedByID string `json:"fallback_created_by_id,omitempty"`

	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithLifecycleContext = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		Operation: "install-sync",
		OwnerID:   s.AppID,
		OwnerType: "apps",
		OwnerName: s.AppName,
		Metadata: map[string]any{
			"app_name":     s.AppName,
			"triggered_by": s.TriggeredBy,
			"commit_sha":   s.CommitSHA,
		},
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	return nil
}
