package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

type CleanupInstallTokensRequest struct {
	InstallID string `validate:"required"`
}

type CleanupInstallTokensResponse struct {
	JobAccounts       int `json:"job_accounts"`
	LogStreamAccounts int `json:"log_stream_accounts"`
	ExpiredTokens     int `json:"expired_tokens"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @max-retries 3
func (a *Activities) CleanupInstallTokens(ctx context.Context, req CleanupInstallTokensRequest) (*CleanupInstallTokensResponse, error) {
	out := &CleanupInstallTokensResponse{}
	streamIDs := map[string]struct{}{}

	var groups []app.RunnerGroup
	if err := a.db.WithContext(ctx).
		Where(app.RunnerGroup{OwnerID: req.InstallID, Type: app.RunnerGroupTypeInstall}).
		Find(&groups).Error; err != nil {
		return nil, fmt.Errorf("load runner groups: %w", err)
	}
	if len(groups) > 0 {
		groupIDs := make([]string, 0, len(groups))
		for _, group := range groups {
			groupIDs = append(groupIDs, group.ID)
		}
		var runners []app.Runner
		if err := a.db.WithContext(ctx).Where("runner_group_id IN ?", groupIDs).Find(&runners).Error; err != nil {
			return nil, fmt.Errorf("load runners: %w", err)
		}
		if len(runners) > 0 {
			runnerIDs := make([]string, 0, len(runners))
			for _, runner := range runners {
				runnerIDs = append(runnerIDs, runner.ID)
			}
			var jobs []app.RunnerJob
			if err := a.db.WithContext(ctx).
				Select("id", "status", "log_stream_id").
				Where("runner_id IN ?", runnerIDs).
				Find(&jobs).Error; err != nil {
				return nil, fmt.Errorf("load jobs: %w", err)
			}
			for _, job := range jobs {
				if job.Status.IsTerminal() {
					if err := account.RevokeDerivedCredentials(a.db.WithContext(ctx), app.TokenSourceTypeRunnerJob, []string{job.ID}); err != nil {
						return nil, fmt.Errorf("revoke job tokens: %w", err)
					}
					if err := a.acctClient.DeleteServiceAccount(ctx, job.ID); err != nil {
						return nil, fmt.Errorf("delete job service account: %w", err)
					}
					out.JobAccounts++
				}
				if job.LogStreamID != nil && *job.LogStreamID != "" {
					streamIDs[*job.LogStreamID] = struct{}{}
				}
			}
			if err := a.deleteExpiredRunnerTokens(ctx, runnerIDs, out); err != nil {
				return nil, err
			}
		}
	}

	if err := a.collectInstallLogStreamIDs(ctx, req.InstallID, streamIDs); err != nil {
		return nil, err
	}
	if err := a.revokeClosedLogStreams(ctx, streamIDs, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (a *Activities) deleteExpiredRunnerTokens(ctx context.Context, runnerIDs []string, out *CleanupInstallTokensResponse) error {
	var accounts []app.Account
	if err := a.db.WithContext(ctx).
		Select("id").
		Where("subject IN ? AND account_type = ?", runnerIDs, app.AccountTypeService).
		Find(&accounts).Error; err != nil {
		return fmt.Errorf("load runner accounts: %w", err)
	}
	if len(accounts) == 0 {
		return nil
	}
	accountIDs := make([]string, 0, len(accounts))
	for _, acct := range accounts {
		accountIDs = append(accountIDs, acct.ID)
	}
	res := a.db.WithContext(ctx).
		Where("account_id IN ? AND expires_at < ?", accountIDs, time.Now()).
		Delete(&app.Token{})
	if res.Error != nil {
		return fmt.Errorf("delete expired runner tokens: %w", res.Error)
	}
	out.ExpiredTokens = int(res.RowsAffected)
	return nil
}

func (a *Activities) collectInstallLogStreamIDs(ctx context.Context, installID string, streamIDs map[string]struct{}) error {
	deployTable := plugins.TableName(a.db, app.InstallDeploy{})
	componentTable := plugins.TableName(a.db, app.InstallComponent{})
	var deployIDs []string
	if err := a.db.WithContext(ctx).Unscoped().Model(&app.InstallDeploy{}).
		Joins("JOIN "+componentTable+" ON "+componentTable+".id = "+deployTable+".install_component_id").
		Where(componentTable+".install_id = ?", installID).
		Pluck(deployTable+".id", &deployIDs).Error; err != nil {
		return fmt.Errorf("load deploys: %w", err)
	}
	if err := a.appendOwnerLogStreams(ctx, "install_deploys", deployIDs, streamIDs); err != nil {
		return err
	}

	var sandboxIDs []string
	if err := a.db.WithContext(ctx).Unscoped().Model(&app.InstallSandboxRun{}).
		Where(app.InstallSandboxRun{InstallID: installID}).
		Pluck("id", &sandboxIDs).Error; err != nil {
		return fmt.Errorf("load sandbox runs: %w", err)
	}
	if err := a.appendOwnerLogStreams(ctx, "install_sandbox_runs", sandboxIDs, streamIDs); err != nil {
		return err
	}

	var actionIDs []string
	if err := a.db.WithContext(ctx).Unscoped().Model(&app.InstallActionWorkflowRun{}).
		Where(app.InstallActionWorkflowRun{InstallID: installID}).
		Pluck("id", &actionIDs).Error; err != nil {
		return fmt.Errorf("load action workflow runs: %w", err)
	}
	if err := a.appendOwnerLogStreams(ctx, "install_action_workflow_runs", actionIDs, streamIDs); err != nil {
		return err
	}

	var workflowIDs []string
	if err := a.db.WithContext(ctx).Unscoped().Model(&app.Workflow{}).
		Where(app.Workflow{OwnerID: installID, OwnerType: plugins.TableName(a.db, app.Install{})}).
		Pluck("id", &workflowIDs).Error; err != nil {
		return fmt.Errorf("load workflows: %w", err)
	}
	if len(workflowIDs) == 0 {
		return nil
	}
	var stepIDs []string
	if err := a.db.WithContext(ctx).Unscoped().Model(&app.WorkflowStep{}).
		Where("install_workflow_id IN ?", workflowIDs).
		Pluck("id", &stepIDs).Error; err != nil {
		return fmt.Errorf("load workflow steps: %w", err)
	}
	var targetIDs []string
	if err := a.db.WithContext(ctx).Unscoped().Model(&app.WorkflowStep{}).
		Where("install_workflow_id IN ? AND step_target_type = ?", workflowIDs, "install_workflow_steps").
		Pluck("step_target_id", &targetIDs).Error; err != nil {
		return fmt.Errorf("load workflow step targets: %w", err)
	}
	stepIDs = append(stepIDs, targetIDs...)
	return a.appendOwnerLogStreams(ctx, "install_workflow_steps", stepIDs, streamIDs)
}

func (a *Activities) appendOwnerLogStreams(ctx context.Context, ownerType string, ownerIDs []string, streamIDs map[string]struct{}) error {
	if len(ownerIDs) == 0 {
		return nil
	}
	var ids []string
	if err := a.db.WithContext(ctx).Unscoped().Model(&app.LogStream{}).
		Where("owner_type = ? AND owner_id IN ?", ownerType, ownerIDs).
		Pluck("id", &ids).Error; err != nil {
		return fmt.Errorf("load %s log streams: %w", ownerType, err)
	}
	for _, id := range ids {
		if id != "" {
			streamIDs[id] = struct{}{}
		}
	}
	return nil
}

func (a *Activities) revokeClosedLogStreams(ctx context.Context, streamIDs map[string]struct{}, out *CleanupInstallTokensResponse) error {
	if len(streamIDs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(streamIDs))
	for id := range streamIDs {
		ids = append(ids, id)
	}
	var streams []app.LogStream
	if err := a.db.WithContext(ctx).Unscoped().Where("id IN ?", ids).Find(&streams).Error; err != nil {
		return fmt.Errorf("load log streams: %w", err)
	}
	for _, stream := range streams {
		if stream.Open && stream.DeletedAt == 0 {
			continue
		}
		if err := a.acctClient.DeleteServiceAccount(ctx, stream.ID); err != nil {
			return fmt.Errorf("delete log stream service account: %w", err)
		}
		out.LogStreamAccounts++
	}
	return nil
}
