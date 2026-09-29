package customstacks

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "sync_custom_stacks"

type Signal struct {
	AppStackConfigID string `json:"app_stack_config_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppStackConfigID == "" {
		return errors.New("app_stack_config_id is required")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	return activities.AwaitUploadCustomNestedStackTemplates(ctx, &activities.UploadCustomNestedStackTemplatesRequest{
		AppStackConfigID: s.AppStackConfigID,
	})
}
