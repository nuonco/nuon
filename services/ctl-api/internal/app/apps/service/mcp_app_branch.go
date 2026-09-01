package service

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

func (s *service) requireAppBranches(ctx context.Context) error {
	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		return err
	}
	if !enabled {
		return features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches)
	}
	return nil
}

func (s *service) findAppRef(ctx context.Context, orgID, appRef string) (*app.App, error) {
	var a app.App
	res := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where(s.db.Where("name = ?", appRef).Or("id = ?", appRef)).
		First(&a)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find app %q: %w", appRef, res.Error)
	}
	return &a, nil
}

func (s *service) findAppBranch(ctx context.Context, orgID, appID, branchRef string) (*app.AppBranch, error) {
	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Where(app.AppBranch{OrgID: orgID, AppID: appID}).
		Where(s.db.Where("name = ?", branchRef).Or("id = ?", branchRef)).
		First(&branch)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find app branch %q: %w", branchRef, res.Error)
	}
	return &branch, nil
}

func (s *service) findInstallOnApp(ctx context.Context, orgID, appID, installRef string) (*app.Install, error) {
	var inst app.Install
	res := s.db.WithContext(ctx).
		Where(app.Install{OrgID: orgID, AppID: appID}).
		Where(s.db.Where("name = ?", installRef).Or("id = ?", installRef)).
		First(&inst)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to find install %q: %w", installRef, res.Error)
	}
	return &inst, nil
}

func (s *service) latestAppBranchConfig(ctx context.Context, branchID string) (*app.AppBranchConfig, error) {
	var cfg app.AppBranchConfig
	res := s.db.WithContext(ctx).
		Preload("PublicGitVCSConfig").
		Preload("ConnectedGithubVCSConfig").
		Preload("InstallGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("app_branch_id = ?", branchID).
		Order("config_number DESC").
		First(&cfg)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest branch config: %w", res.Error)
	}
	return &cfg, nil
}

func (s *service) latestAppBranchRun(ctx context.Context, branchID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	res := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Preload("Preview").
		Preload("Comparison").
		Preload("Comparison.BaseRun").
		Preload("Comparison.BaseRun.VCSConnectionCommit").
		Where("app_branch_id = ?", branchID).
		Order("created_at DESC").
		First(&run)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest branch run: %w", res.Error)
	}
	return &run, nil
}

func (s *service) markRunAwaitingApproval(ctx context.Context, run *app.AppBranchRun) error {
	if run == nil || run.WorkflowID == nil {
		return nil
	}

	var count int64
	res := s.db.WithContext(ctx).
		Model(&app.WorkflowStep{}).
		Joins("JOIN install_workflow_step_approvals approvals "+
			"ON approvals.install_workflow_step_id = install_workflow_steps.id AND approvals.deleted_at = 0").
		Joins("LEFT JOIN install_workflow_step_approval_responses responses "+
			"ON responses.install_workflow_step_approval_id = approvals.id AND responses.deleted_at = 0").
		Where("install_workflow_steps.install_workflow_id = ?", *run.WorkflowID).
		Where("install_workflow_steps.execution_type = ?", app.WorkflowStepExecutionTypeApproval).
		Where("install_workflow_steps.status->>'status' = ?", string(app.AwaitingApproval)).
		Where("responses.id IS NULL").
		Count(&count)
	if res.Error != nil {
		return res.Error
	}
	run.AwaitingApproval = count > 0
	return nil
}

func mcpBranchVCS(cfg *app.AppBranchConfig) *mcpAppBranchVCS {
	if cfg == nil {
		return nil
	}
	if gh := cfg.ConnectedGithubVCSConfig; gh != nil {
		repo := gh.Repo
		if gh.RepoOwner != "" && gh.RepoName != "" {
			repo = gh.RepoOwner + "/" + gh.RepoName
		}
		return &mcpAppBranchVCS{Repo: repo, GitBranch: gh.Branch, Directory: gh.Directory}
	}
	if pub := cfg.PublicGitVCSConfig; pub != nil {
		return &mcpAppBranchVCS{Repo: pub.Repo, GitBranch: pub.Branch, Directory: pub.Directory}
	}
	return nil
}

type mcpAppBranchVCS struct {
	Repo      string `json:"repo,omitempty"`
	GitBranch string `json:"git_branch,omitempty"`
	Directory string `json:"directory,omitempty"`
}
