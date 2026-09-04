package addinstall

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	SignalType signal.SignalType = "app-branch-add-install"
	queueName                    = "app-branch-signals"
)

type Signal struct {
	AppBranchID    string `json:"app_branch_id"`
	InstallID      string `json:"install_id"`
	InstallGroupID string `json:"install_group_id,omitempty"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType { return SignalType }

func (s *Signal) Validate(_ workflow.Context) error {
	if s.AppBranchID == "" {
		return fmt.Errorf("app_branch_id is required")
	}
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	result, err := activities.AwaitGetLatestActiveBranchAppConfig(ctx, &activities.GetLatestActiveBranchAppConfigInput{
		AppBranchID: s.AppBranchID,
		InstallID:   s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to resolve branch app config: %w", err)
	}
	if result.AlreadyCurrent || result.AppConfigID == "" {
		return nil
	}

	_, err = activities.AwaitCreateInstallAppConfigVersionWorkflow(ctx, &activities.CreateInstallAppConfigVersionWorkflowInput{
		InstallID:      s.InstallID,
		NewAppConfigID: result.AppConfigID,
		InstallGroupID: s.InstallGroupID,
	})
	if err != nil {
		return fmt.Errorf("unable to apply branch app config: %w", err)
	}
	return nil
}

func Enqueue(ctx context.Context, client *queueclient.Client, appBranchID, installID, installGroupID string) error {
	queue, err := client.GetQueueByOwnerAndName(ctx, appBranchID, "app_branches", queueName)
	if err != nil {
		return fmt.Errorf("unable to find app branch queue: %w", err)
	}
	_, err = client.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &Signal{
			AppBranchID:    appBranchID,
			InstallID:      installID,
			InstallGroupID: installGroupID,
		},
	})
	return err
}
