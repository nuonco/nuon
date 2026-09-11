package activities

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
)

type TriggerAppBranchRunFromCreatedRequest struct {
	AppBranchID       string `json:"app_branch_id" validate:"required"`
	AppBranchConfigID string `json:"app_branch_config_id" validate:"required"`
}

type TriggerAppBranchRunFromCreatedResponse struct {
	RunID      string `json:"run_id,omitempty"`
	WorkflowID string `json:"workflow_id,omitempty"`
	Skipped    bool   `json:"skipped,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) TriggerAppBranchRunFromCreated(ctx context.Context, req TriggerAppBranchRunFromCreatedRequest) (*TriggerAppBranchRunFromCreatedResponse, error) {
	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Preload("Queue").First(&branch, "id = ?", req.AppBranchID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch: %w", err)
	}

	var config app.AppBranchConfig
	if err := a.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		First(&config, "id = ?", req.AppBranchConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch config: %w", err)
	}

	var runCount int64
	if err := a.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{AppBranchID: req.AppBranchID}).
		Count(&runCount).Error; err != nil {
		return nil, fmt.Errorf("unable to count app branch runs: %w", err)
	}

	if skip, reason := skipFirstAppBranchRun(runCount > 0, config.ConnectedGithubVCSConfig != nil, config.PublicGitVCSConfig != nil); skip {
		return &TriggerAppBranchRunFromCreatedResponse{Skipped: true, Reason: reason}, nil
	}

	if branch.Queue.ID == "" {
		return nil, fmt.Errorf("app branch %s has no queue", req.AppBranchID)
	}

	metadata := map[string]string{
		"app_id":        branch.AppID,
		"config_id":     config.ID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"event_type":    "app_branch_created",
		"run_type":      string(app.AppBranchRunTypeManual),
	}

	triggerResp, err := a.helpers.TriggerAppBranchRun(ctx, &appshelpers.TriggerAppBranchRunRequest{
		Run: appshelpers.CreateAppBranchRunRequest{
			AppBranchID:       req.AppBranchID,
			AppBranchConfigID: req.AppBranchConfigID,
			RunType:           app.AppBranchRunTypeManual,
			EventType:         "app_branch_created",
		},
		QueueID:        branch.Queue.ID,
		Metadata:       metadata,
		ApprovalOption: app.InstallApprovalOptionPrompt,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to trigger app branch run: %w", err)
	}

	return &TriggerAppBranchRunFromCreatedResponse{
		RunID:      triggerResp.Run.ID,
		WorkflowID: triggerResp.Workflow.ID,
	}, nil
}

func skipFirstAppBranchRun(hasExistingRun, hasConnectedRepo, hasPublicRepo bool) (bool, string) {
	if hasExistingRun {
		return true, "run_exists"
	}
	if !hasConnectedRepo && !hasPublicRepo {
		return true, "no_vcs"
	}
	return false, ""
}
