package customstacks

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// SignalType triggers uploading an app stack config's custom nested stack
// templates to the managed S3 bucket.
const SignalType signal.SignalType = "sync_custom_stacks"

// Signal uploads the custom nested stack template contents for an app stack
// config to S3, setting each stack's ContentsHash and marking it ready.
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
