package labeladded

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/addinstall"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
)

const SignalType signal.SignalType = "label-added"

type Signal struct {
	InstallID string `json:"install_id"`
	LabelName string `json:"label_name"`
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithAutoRetry        = (*Signal)(nil)
	_ signal.SignalWithMaxRetries       = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType { return SignalType }
func (s *Signal) AutoRetry() bool         { return true }
func (s *Signal) MaxRetries() int         { return 5 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: "label-added",
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		Metadata: map[string]any{
			"label_name": s.LabelName,
		},
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if s.LabelName == "" {
		return fmt.Errorf("label_name is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	match, err := activities.AwaitReconcileInstallBranchForLabels(ctx, &activities.ReconcileInstallBranchForLabelsInput{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to reconcile app branch membership: %w", err)
	}
	if match.AppBranchID == "" {
		return nil
	}

	_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
		OwnerID:   match.AppBranchID,
		OwnerType: "app_branches",
		QueueName: "app-branch-signals",
		Signal: &addinstall.Signal{
			AppBranchID:    match.AppBranchID,
			InstallID:      s.InstallID,
			InstallGroupID: match.InstallGroupID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to enqueue app branch install update: %w", err)
	}
	return nil
}
