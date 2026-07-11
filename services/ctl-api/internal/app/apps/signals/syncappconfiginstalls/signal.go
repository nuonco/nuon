package syncappconfiginstalls

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updateappconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "sync-app-config-installs"

type Signal struct {
	AppID          string `json:"app_id" validate:"required"`
	NewAppConfigID string `json:"new_app_config_id" validate:"required"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		OwnerID:   s.AppID,
		OwnerType: "apps",
		Operation: "sync-app-config-installs",
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.AppID == "" {
		return fmt.Errorf("app_id is required")
	}
	if s.NewAppConfigID == "" {
		return fmt.Errorf("new_app_config_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	result, err := AwaitGetNonBranchManagedInstallIDs(ctx, &GetNonBranchManagedInstallIDsInput{
		AppID: s.AppID,
	})
	if err != nil {
		return fmt.Errorf("unable to get non-branch-managed installs: %w", err)
	}

	if len(result.InstallIDs) == 0 {
		logger.Info("no non-branch-managed installs to update")
		return nil
	}

	logger.Info("syncing app config to installs",
		"install_count", len(result.InstallIDs),
		"new_app_config_id", s.NewAppConfigID,
	)

	for _, installID := range result.InstallIDs {
		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   installID,
			OwnerType: "installs",
			QueueName: "install-signals",
			Signal: &updateappconfig.Signal{
				InstallID:      installID,
				NewAppConfigID: s.NewAppConfigID,
				TriggeredBy:    "sync",
				Metadata:       map[string]string{"source": "sync"},
			},
		}); err != nil {
			logger.Warn("unable to enqueue update-app-config signal",
				"install_id", installID,
				"error", err,
			)
		}
	}

	return nil
}
