package ignorechanges

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-ignore-changes"

type Signal struct {
	RunID        string   `json:"run_id" validate:"required"`
	AppBranchID  string   `json:"app_branch_id" validate:"required"`
	BaseSHA      string   `json:"base_sha,omitempty"`
	ChangedFiles []string `json:"changed_files,omitempty"`

	FlowID string `json:"flow_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)
var _ signal.SignalWithStepContext = (*Signal)(nil)

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.StepID = stepID
	s.FlowID = flowID
}

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if err := validator.New().Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}
	_, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	return err
}
