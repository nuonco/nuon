package deploygrouptoqueue

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-deploy-group-to-queue"

type Signal struct {
	InstallGroupID string `json:"install_group_id" validate:"required"`
	AppBranchID    string `json:"app_branch_id" validate:"required"`
	RunID          string `json:"run_id" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	// Use playground validator for struct tag validation
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	// Validate app branch exists
	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	// Validate install group exists
	_, err = activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return errors.Wrap(err, "install group not found")
	}

	return nil
}
