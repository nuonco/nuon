package activities

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
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
	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Preload("Queue").First(&branch, "id = ?", appBranchID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch: %w", err)
	}
	if branch.Queue.ID == "" {
		return nil, fmt.Errorf("app branch %s has no queue", appBranchID)
	}

	var config app.AppBranchConfig
	if err := a.db.WithContext(ctx).First(&config, "id = ?", appBranchConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch config: %w", err)
	}

	triggerResp, err := a.appsHelpers.TriggerAppBranchRun(ctx, &appshelpers.TriggerAppBranchRunRequest{
		Run: appshelpers.CreateAppBranchRunRequest{
			AppBranchID:       appBranchID,
			AppBranchConfigID: appBranchConfigID,
			Force:             true,
		},
		QueueID: branch.Queue.ID,
		Metadata: map[string]string{
			"config_id":     appBranchConfigID,
			"config_number": strconv.Itoa(config.ConfigNumber),
			"force":         "true",
			"event_type":    "onboarding",
		},
		Callback: cb,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to trigger app branch run: %w", err)
	}

	return &TriggerOnboardingAppBranchRunResponse{
		RunID:         triggerResp.Run.ID,
		WorkflowID:    triggerResp.Workflow.ID,
		QueueSignalID: triggerResp.QueueSignalID,
	}, nil
}
