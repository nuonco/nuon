package buildcomponents

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
)

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
