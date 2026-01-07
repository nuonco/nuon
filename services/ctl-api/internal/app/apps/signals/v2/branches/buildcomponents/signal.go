package buildcomponents

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-build-components"

type Signal struct {
	AppBranchID string `json:"app_branch_id"`
	AppConfigID string `json:"app_config_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppBranchID == "" {
		return errors.New("app_branch_id is required")
	}
	if s.AppConfigID == "" {
		return errors.New("app_config_id is required")
	}

	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Get app config with components
	appConfig, err := activities.AwaitGetAppConfigByIDByAppConfigID(ctx, s.AppConfigID)
	if err != nil {
		return fmt.Errorf("unable to get app config: %w", err)
	}

	l.Info("triggering component builds",
		"app_branch_id", s.AppBranchID,
		"app_config_id", s.AppConfigID,
		"component_count", len(appConfig.ComponentIDs))

	// Trigger builds for each component
	for _, componentID := range appConfig.ComponentIDs {
		if err := activities.AwaitTriggerComponentBuildByComponentID(ctx, componentID); err != nil {
			// Log error but continue with other components
			l.Error("unable to trigger component build",
				zap.String("component_id", componentID),
				zap.Error(err))
		}
	}

	return nil
}
