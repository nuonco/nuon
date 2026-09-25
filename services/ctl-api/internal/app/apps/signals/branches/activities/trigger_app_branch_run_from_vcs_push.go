package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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
	AppBranchID         string   `json:"app_branch_id"`
	AppBranchConfigID   string   `json:"app_branch_config_id"`
	PlanOnly            bool     `json:"plan_only,omitempty"`
	EventType           string   `json:"event_type,omitempty"`
	PRNumber            *int     `json:"pr_number,omitempty"`
	HeadSHA             string   `json:"head_sha,omitempty"`
	HeadRef             string   `json:"head_ref,omitempty"`
	BaseBranch          string   `json:"base_branch,omitempty"`
	BaseSHA             string   `json:"base_sha,omitempty"`
	ChangedFiles        []string `json:"changed_files,omitempty"`
	PusherEmails        []string `json:"pusher_emails,omitempty"`
	SenderLogin         string   `json:"sender_login,omitempty"`
	FallbackCreatedByID string   `json:"fallback_created_by_id,omitempty"`
	Draft               bool     `json:"draft,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
func (a *Activities) TriggerAppBranchRunFromVCSPush(ctx context.Context, req TriggerAppBranchRunFromVCSPushRequest) (*TriggerAppBranchRunFromVCSPushResponse, error) {
	appBranchID := req.AppBranchID
	appBranchConfigID := req.AppBranchConfigID

	var branch app.AppBranch
	if err := a.db.WithContext(ctx).Preload("Queue", app.DefaultQueueScope).First(&branch, "id = ?", appBranchID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch: %w", err)
	}

	if branch.Queue.ID == "" {
		return nil, fmt.Errorf("app branch %s has no queue", appBranchID)
	}

	var config app.AppBranchConfig
	if err := a.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig.VCSConnection").
		Preload("PublicGitVCSConfig").
		First(&config, "id = ?", appBranchConfigID).Error; err != nil {
		return nil, fmt.Errorf("unable to find app branch config: %w", err)
	}

	runMetadata, shouldRun, err := a.evaluateBranchRunConfig(ctx, &config, &req)
	if err != nil {
		return nil, err
	}
	if !shouldRun {
		a.l.Info("skipping VCS event due to app branch run config",
			zap.String("app_branch_id", appBranchID),
			zap.String("event_type", req.EventType),
		)
		return &TriggerAppBranchRunFromVCSPushResponse{}, nil
	}

	runType := RunTypeFromEventType(req.EventType)
	previewDefaults := appshelpers.BranchPreviewConfigOrDefault(&config)
	if runType == app.AppBranchRunTypeGitPreview && previewDefaults.Mode == app.AppBranchRunPreviewModeNone {
		a.l.Info("skipping pull request because previews are disabled",
			zap.String("app_branch_id", appBranchID),
			zap.String("app_branch_config_id", appBranchConfigID),
		)
		return &TriggerAppBranchRunFromVCSPushResponse{}, nil
	}
	if runType == app.AppBranchRunTypeGitPreview &&
		previewDefaults.Mode != app.AppBranchRunPreviewModeBuildOnly &&
		!previewDefaults.HasInstallTarget() {
		a.l.Info("skipping pull request because preview has no install target",
			zap.String("app_branch_id", appBranchID),
			zap.String("app_branch_config_id", appBranchConfigID),
			zap.String("preview_mode", string(previewDefaults.Mode)),
		)
		return &TriggerAppBranchRunFromVCSPushResponse{}, nil
	}
	if req.Draft && previewDefaults.IgnoreDrafts {
		a.l.Info("skipping draft pull request preview",
			zap.String("app_branch_id", appBranchID),
			zap.String("app_branch_config_id", appBranchConfigID),
		)
		return &TriggerAppBranchRunFromVCSPushResponse{}, nil
	}

	gitRef := req.HeadRef
	if gitRef == "" {
		gitRef = req.HeadSHA
	}

	ctx = a.resolvePusherAccount(ctx, branch.OrgID, req.PusherEmails, req.FallbackCreatedByID)

	runLabels := BuildRunLabels(&req)

	metadata := map[string]string{
		"app_id":        branch.AppID,
		"config_id":     appBranchConfigID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"force":         "false",
		"event_type":    string(runMetadata.Trigger),
		"run_type":      string(runType),
	}
	if runMetadata.Tag != "" {
		metadata["tag"] = runMetadata.Tag
	}
	if runMetadata.GithubLabel != "" {
		metadata["github_label"] = runMetadata.GithubLabel
	}
	if req.PRNumber != nil {
		metadata["pr_number"] = strconv.Itoa(*req.PRNumber)
	}
	if req.HeadSHA != "" {
		metadata["head_sha"] = req.HeadSHA
	}
	if gitRef != "" {
		metadata["git_ref"] = gitRef
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
			GitRef:            gitRef,
			BaseBranch:        req.BaseBranch,
			IsDraftMode:       req.Draft,
			Metadata:          runMetadata,
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

func (a *Activities) evaluateBranchRunConfig(ctx context.Context, config *app.AppBranchConfig, req *TriggerAppBranchRunFromVCSPushRequest) (app.AppBranchRunMetadata, bool, error) {
	runConfig := app.AppBranchRunConfig{Mode: app.AppBranchRunModePush}
	if config.RunConfig != nil {
		runConfig = *config.RunConfig
		runConfig.Normalize()
	}

	metadata := app.AppBranchRunMetadata{
		Trigger:    app.AppBranchRunTrigger(req.EventType),
		HeadSHA:    req.HeadSHA,
		GitRef:     req.HeadRef,
		BaseBranch: req.BaseBranch,
		PRNumber:   req.PRNumber,
		IsDraft:    req.Draft,
		RunMode:    string(runConfig.Mode),
		TagPrefix:  runConfig.TagPrefix,
	}

	switch req.EventType {
	case "pull_request":
		return metadata, runConfig.Mode == app.AppBranchRunModePush, nil
	case "tag":
		if runConfig.Mode != app.AppBranchRunModeTagPrefix || !strings.HasPrefix(req.HeadRef, runConfig.TagPrefix) {
			return metadata, false, nil
		}
		metadata.Tag = req.HeadRef
		baseBranch, vcsConfigID, err := branchVCSIdentity(config)
		if err != nil {
			return metadata, false, err
		}
		metadata.BaseBranch = baseBranch
		owner, repo, client, err := a.resolveGithubClient(ctx, vcsConfigID)
		if err != nil {
			return metadata, false, err
		}
		commit, _, err := client.Repositories.GetCommit(ctx, owner, repo, req.HeadRef, nil)
		if err != nil {
			return metadata, false, fmt.Errorf("unable to resolve tag to commit: %w", err)
		}
		req.HeadSHA = commit.GetSHA()
		metadata.HeadSHA = req.HeadSHA
		comparison, _, err := client.Repositories.CompareCommits(ctx, owner, repo, baseBranch, req.HeadSHA, nil)
		if err != nil {
			return metadata, false, fmt.Errorf("unable to verify tagged commit ancestry: %w", err)
		}
		status := comparison.GetStatus()
		return metadata, status == "behind" || status == "identical", nil
	case "push":
		switch runConfig.Mode {
		case app.AppBranchRunModePush:
			return metadata, true, nil
		case app.AppBranchRunModeGithubLabel:
			baseBranch, vcsConfigID, err := branchVCSIdentity(config)
			if err != nil {
				return metadata, false, err
			}
			owner, repo, client, err := a.resolveGithubClient(ctx, vcsConfigID)
			if err != nil {
				return metadata, false, err
			}
			prs, _, err := client.PullRequests.ListPullRequestsWithCommit(ctx, owner, repo, req.HeadSHA, nil)
			if err != nil {
				return metadata, false, fmt.Errorf("unable to list pull requests for commit: %w", err)
			}
			for _, pr := range prs {
				for _, label := range pr.Labels {
					if label.GetName() != runConfig.GithubLabel {
						continue
					}
					metadata.Trigger = app.AppBranchRunTriggerGithubLabel
					metadata.BaseBranch = baseBranch
					metadata.PRNumber = pr.Number
					metadata.GithubLabel = runConfig.GithubLabel
					req.PRNumber = pr.Number
					req.BaseBranch = baseBranch
					return metadata, true, nil
				}
			}
			return metadata, false, nil
		default:
			return metadata, false, nil
		}
	default:
		return metadata, false, nil
	}
}

func branchVCSIdentity(config *app.AppBranchConfig) (string, string, error) {
	if config.ConnectedGithubVCSConfig != nil {
		return config.ConnectedGithubVCSConfig.Branch, config.ConnectedGithubVCSConfig.ID, nil
	}
	if config.PublicGitVCSConfig != nil {
		return config.PublicGitVCSConfig.Branch, config.PublicGitVCSConfig.ID, nil
	}
	return "", "", fmt.Errorf("app branch config has no VCS config")
}

func (a *Activities) resolvePusherAccount(ctx context.Context, orgID string, emails []string, fallbackCreatedByID string) context.Context {
	for _, email := range emails {
		if email == "" {
			continue
		}
		var account app.Account
		err := a.db.WithContext(ctx).
			// emails are stored lowercased; compare on the raw column so idx_accounts_email is used
			Where("accounts.email = LOWER(?)", email).
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
