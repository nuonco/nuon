package activities

import (
	"context"
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// appBranchRunSignal is a minimal signal type that matches the
// run.Signal type string. We define it here to avoid an import cycle
// (activities cannot import branches/run which imports activities).
type appBranchRunSignal struct {
	RunID string `json:"run_id"`
}

func (s *appBranchRunSignal) Type() signal.SignalType           { return "app-branch-run" }
func (s *appBranchRunSignal) Validate(_ workflow.Context) error { return nil }
func (s *appBranchRunSignal) Execute(_ workflow.Context) error  { return nil }

type TriggerAppBranchRunFromVCSPushResponse struct {
	RunID         string `json:"run_id"`
	WorkflowID    string `json:"workflow_id"`
	QueueSignalID string `json:"queue_signal_id"`
}

type TriggerAppBranchRunFromVCSPushRequest struct {
	AppBranchID       string `json:"app_branch_id"`
	AppBranchConfigID string `json:"app_branch_config_id"`
	PlanOnly          bool   `json:"plan_only,omitempty"`
	EventType         string `json:"event_type,omitempty"`
	PRNumber          *int   `json:"pr_number,omitempty"`
	HeadSHA           string `json:"head_sha,omitempty"`
	BaseBranch        string `json:"base_branch,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) TriggerAppBranchRunFromVCSPush(ctx context.Context, req TriggerAppBranchRunFromVCSPushRequest) (*TriggerAppBranchRunFromVCSPushResponse, error) {
	appBranchID := req.AppBranchID
	appBranchConfigID := req.AppBranchConfigID
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
	run, err := a.helpers.CreateAppBranchRun(ctx, &appshelpers.CreateAppBranchRunRequest{
		AppBranchID:       appBranchID,
		AppBranchConfigID: appBranchConfigID,
		Force:             false,
		PlanOnly:          req.PlanOnly,
		EventType:         req.EventType,
		PRNumber:          req.PRNumber,
		HeadSHA:           req.HeadSHA,
		BaseBranch:        req.BaseBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create app branch run: %w", err)
	}

	// Create workflow
	metadata := map[string]string{
		"run_id":        run.ID,
		"app_id":        branch.AppID,
		"config_id":     appBranchConfigID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"force":         "false",
		"event_type":    req.EventType,
		"commit_sha":    run.CommitSHA,
	}
	if req.PRNumber != nil {
		metadata["pr_number"] = strconv.Itoa(*req.PRNumber)
	}
	if req.HeadSHA != "" {
		metadata["head_sha"] = req.HeadSHA
	}
	if req.BaseBranch != "" {
		metadata["base_branch"] = req.BaseBranch
	}

	wf, err := a.helpers.CreateWorkflow(
		ctx,
		appBranchID,
		app.WorkflowTypeAppBranchesRun,
		metadata,
		req.PlanOnly,
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
		Signal: &appBranchRunSignal{
			RunID: run.ID,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to enqueue run signal: %w", err)
	}

	return &TriggerAppBranchRunFromVCSPushResponse{
		RunID:         run.ID,
		WorkflowID:    wf.ID,
		QueueSignalID: enqueueResp.ID,
	}, nil
}
