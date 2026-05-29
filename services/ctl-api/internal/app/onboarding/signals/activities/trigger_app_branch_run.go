package activities

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	runsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/run"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type TriggerOnboardingAppBranchRunResponse struct {
	RunID         string `json:"run_id"`
	WorkflowID    string `json:"workflow_id"`
	QueueSignalID string `json:"queue_signal_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) triggerOnboardingAppBranchRun(ctx context.Context, appBranchID, appBranchConfigID string, cb callback.Ref) (*TriggerOnboardingAppBranchRunResponse, error) {
	// Load branch with queue
	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Preload("Queue").First(&branch, "id = ?", appBranchID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch: %w", err)
	}

	if branch.Queue.ID == "" {
		return nil, fmt.Errorf("app branch %s has no queue", appBranchID)
	}

	// Load config to get config number
	var config app.AppBranchConfig
	if err := a.db.WithContext(ctx).First(&config, "id = ?", appBranchConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch config: %w", err)
	}

	// Create app branch run
	run, err := a.appsHelpers.CreateAppBranchRun(ctx, &appshelpers.CreateAppBranchRunRequest{
		AppBranchID:       appBranchID,
		AppBranchConfigID: appBranchConfigID,
		Force:             true,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create app branch run: %w", err)
	}

	// Create workflow
	wf, err := a.appsHelpers.CreateWorkflow(
		ctx,
		appBranchID,
		app.WorkflowTypeAppBranchesRun,
		map[string]string{
			"run_id":        run.ID,
			"config_id":     appBranchConfigID,
			"config_number": strconv.Itoa(config.ConfigNumber),
			"force":         "true",
		},
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create workflow: %w", err)
	}

	// Update run with workflow ID
	run.WorkflowID = &wf.ID
	if err := a.db.WithContext(ctx).Save(run).Error; err != nil {
		return nil, fmt.Errorf("unable to update run with workflow id: %w", err)
	}

	// Enqueue run signal on the branch queue
	enqueueResp, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   branch.Queue.ID,
		OwnerID:   run.ID,
		OwnerType: plugins.TableName(a.db, app.AppBranchRun{}),
		Signal: &runsignal.Signal{
			RunID: run.ID,
		},
		Callback: cb,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to enqueue run signal: %w", err)
	}

	return &TriggerOnboardingAppBranchRunResponse{
		RunID:         run.ID,
		WorkflowID:    wf.ID,
		QueueSignalID: enqueueResp.ID,
	}, nil
}
