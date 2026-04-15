package run

import (
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-run"

type Signal struct {
	signal.Hooks
	RunID string `json:"run_id" validate:"required"` // The app branch run ID - everything else fetched from DB
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

	// Validate run exists and fetch all related data
	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return errors.Wrap(err, "app branch run not found")
	}

	// Validate run has required relationships
	if run.AppBranchID == "" {
		return errors.New("run missing app_branch_id")
	}
	if run.AppBranchConfigID == "" {
		return errors.New("run missing app_branch_config_id")
	}

	return nil
}
