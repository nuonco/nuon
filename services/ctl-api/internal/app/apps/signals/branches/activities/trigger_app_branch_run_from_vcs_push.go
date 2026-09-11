package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

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
	BaseSHA           string   `json:"base_sha,omitempty"`
	ChangedFiles      []string `json:"changed_files,omitempty"`
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

	metadata := map[string]string{
		"app_id":        branch.AppID,
		"config_id":     appBranchConfigID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"force":         "false",
		"event_type":    req.EventType,
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
	if req.BaseSHA != "" {
		metadata["base_sha"] = req.BaseSHA
	}
	if len(req.ChangedFiles) > 0 {
		changedFiles, err := json.Marshal(req.ChangedFiles)
		if err != nil {
			return nil, fmt.Errorf("unable to encode changed files: %w", err)
		}
		metadata["changed_files"] = string(changedFiles)
	}

	triggerResp, err := a.helpers.TriggerAppBranchRun(ctx, &appshelpers.TriggerAppBranchRunRequest{
		Run: appshelpers.CreateAppBranchRunRequest{
			AppBranchID:       appBranchID,
			AppBranchConfigID: appBranchConfigID,
			RunType:           runType,
			PlanOnly:          req.PlanOnly,
			EventType:         req.EventType,
			PRNumber:          req.PRNumber,
			HeadSHA:           req.HeadSHA,
			BaseBranch:        req.BaseBranch,
			Labels:            runLabels,
		},
		QueueID:  branch.Queue.ID,
		Metadata: metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to trigger app branch run: %w", err)
	}

	return &TriggerAppBranchRunFromVCSPushResponse{
		RunID:         triggerResp.Run.ID,
		WorkflowID:    triggerResp.Workflow.ID,
		QueueSignalID: triggerResp.QueueSignalID,
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
