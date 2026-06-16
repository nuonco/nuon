package vcspush

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "vcs-push"

type Signal struct {
	AppBranchID       string `json:"app_branch_id" validate:"required"`
	AppBranchConfigID string `json:"app_branch_config_id" validate:"required"`

	// PR metadata — populated for pull_request events, empty for push events
	PlanOnly   bool   `json:"plan_only,omitempty"`
	EventType  string `json:"event_type,omitempty"` // "push" or "pull_request"
	PRNumber   *int   `json:"pr_number,omitempty"`
	HeadSHA    string `json:"head_sha,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	// Verify app branch exists
	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	return nil
}
