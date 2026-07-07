package activities

import (
	"context"
	"fmt"
	"strconv"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	AppBranchID       string   `json:"app_branch_id"`
	AppBranchConfigID string   `json:"app_branch_config_id"`
	PlanOnly          bool     `json:"plan_only,omitempty"`
	EventType         string   `json:"event_type,omitempty"`
	PRNumber          *int     `json:"pr_number,omitempty"`
	HeadSHA           string   `json:"head_sha,omitempty"`
	BaseBranch        string   `json:"base_branch,omitempty"`
	PusherEmails      []string `json:"pusher_emails,omitempty"`

	SenderLogin         string `json:"sender_login,omitempty"`
	FallbackCreatedByID string `json:"fallback_created_by_id,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) TriggerAppBranchRunFromVCSPush(ctx context.Context, req TriggerAppBranchRunFromVCSPushRequest) (*TriggerAppBranchRunFromVCSPushResponse, error) {
	appBranchID := req.AppBranchID
	appBranchConfigID := req.AppBranchConfigID

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

	ctx = a.resolvePusherAccount(ctx, branch.OrgID, req.PusherEmails, req.FallbackCreatedByID)

	runType := RunTypeFromEventType(req.EventType)
	runLabels := BuildRunLabels(&req)

	run, err := a.helpers.CreateAppBranchRun(ctx, &appshelpers.CreateAppBranchRunRequest{
		AppBranchID:       appBranchID,
		AppBranchConfigID: appBranchConfigID,
		RunType:           runType,
		Force:             false,
		PlanOnly:          req.PlanOnly,
		EventType:         req.EventType,
		PRNumber:          req.PRNumber,
		HeadSHA:           req.HeadSHA,
		BaseBranch:        req.BaseBranch,
		Labels:            runLabels,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create app branch run: %w", err)
	}

	metadata := map[string]string{
		"run_id":        run.ID,
		"app_id":        branch.AppID,
		"config_id":     appBranchConfigID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"force":         "false",
		"event_type":    req.EventType,
		"commit_sha":    run.CommitSHA,
		"run_type":      string(runType),
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

	run.WorkflowID = &wf.ID
	if err := a.db.WithContext(ctx).Save(run).Error; err != nil {
		return nil, fmt.Errorf("unable to update run with workflow id: %w", err)
	}

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

func (a *Activities) resolvePusherAccount(ctx context.Context, orgID string, emails []string, fallbackCreatedByID string) context.Context {
	for _, email := range emails {
		if email == "" {
			continue
		}
		var account app.Account
		err := a.db.WithContext(ctx).
			Where("LOWER(accounts.email) = LOWER(?)", email).
			Joins("JOIN account_roles ON account_roles.account_id = accounts.id").
			Joins("JOIN roles ON roles.id = account_roles.role_id AND roles.org_id = ?", orgID).
			First(&account).Error
		if err == nil {
			a.l.Info("resolved pusher account",
				zap.String("matched_email", email),
				zap.String("account_id", account.ID),
			)
			return cctx.SetAccountIDContext(ctx, account.ID)
		}
	}

	if fallbackCreatedByID != "" {
		return cctx.SetAccountIDContext(ctx, fallbackCreatedByID)
	}

	return ctx
}
