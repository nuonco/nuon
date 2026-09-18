package appbranchchanged

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-changed"

type Signal struct {
	InstallID      string `json:"install_id"`
	AppBranchID    string `json:"app_branch_id"`
	InstallGroupID string `json:"install_group_id,omitempty"`
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
		Operation: "app-branch-changed",
		OwnerID:   s.InstallID,
		OwnerType: "installs",
		Metadata: map[string]any{
			"app_branch_id": s.AppBranchID,
		},
	}
}

func (s *Signal) Validate(_ workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if s.AppBranchID == "" {
		return fmt.Errorf("app_branch_id is required")
	}
	return nil
}

func Enqueue(ctx context.Context, client *queueclient.Client, installID, appBranchID, installGroupID string) error {
	queue, err := client.GetQueueByOwnerAndName(ctx, installID, "installs", "install-signals")
	if err != nil {
		return fmt.Errorf("unable to find install queue: %w", err)
	}
	_, err = client.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		OwnerID:   installID,
		OwnerType: "installs",
		Signal: &Signal{
			InstallID:      installID,
			AppBranchID:    appBranchID,
			InstallGroupID: installGroupID,
		},
	})
	return err
}
