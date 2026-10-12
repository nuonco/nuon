package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallOverviewCommit struct {
	SHA              string     `json:"sha,omitempty"`
	RunID            string     `json:"run_id,omitempty"`
	WorkflowID       string     `json:"workflow_id,omitempty"`
	BranchID         string     `json:"branch_id,omitempty"`
	AppConfigID      string     `json:"app_config_id,omitempty"`
	Message          string     `json:"message,omitempty"`
	Author           string     `json:"author,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	RunStatus        string     `json:"run_status,omitempty"`
	AwaitingApproval bool       `json:"awaiting_approval,omitempty"`
}

type InstallBranchTracking struct {
	BranchID       string                  `json:"branch_id,omitempty"`
	TargetBranch   string                  `json:"target_branch,omitempty"`
	Repo           string                  `json:"repo,omitempty"`
	GitBranch      string                  `json:"git_branch,omitempty"`
	Directory      string                  `json:"directory,omitempty"`
	Status         string                  `json:"status"`
	ExpectedCommit *InstallOverviewCommit  `json:"expected_commit,omitempty"`
	AppliedCommit  *InstallOverviewCommit  `json:"applied_commit,omitempty"`
	SelectedCommit *InstallOverviewCommit  `json:"selected_commit,omitempty"`
	CommitsBehind  *int64                  `json:"commits_behind,omitempty" extensions:"x-nullable"`
	PendingCommits []InstallOverviewCommit `json:"pending_commits"`
}

type InstallConfigDriftResource struct {
	Drifted            bool   `json:"drifted"`
	AppliedAppConfigID string `json:"applied_app_config_id,omitempty"`
}

type InstallConfigDriftComponent struct {
	Name               string `json:"name"`
	ComponentID        string `json:"component_id"`
	Drifted            bool   `json:"drifted"`
	AppliedAppConfigID string `json:"applied_app_config_id,omitempty"`
}

type InstallConfigDrift struct {
	CurrentAppConfigID string                        `json:"current_app_config_id,omitempty"`
	Stack              *InstallConfigDriftResource   `json:"stack,omitempty"`
	Sandbox            *InstallConfigDriftResource   `json:"sandbox,omitempty"`
	Components         []InstallConfigDriftComponent `json:"components"`
}

type InstallOverviewResponse struct {
	BranchTracking InstallBranchTracking `json:"branch_tracking"`
	ConfigDrift    InstallConfigDrift    `json:"config_drift"`
}

// @ID						GetInstallOverview
// @Summary				install overview
// @Description			Returns branch tracking and config drift for an install. commits_behind counts distinct newer app configurations on the tracked branch relative to the install's selected app_config_id, excluding previews and no-config-change runs. It is omitted when selected branch provenance is unknown. pending_commits includes the newest 50 configurations and their latest branch runs. A stack, sandbox, or component is drifted when its applied app config is set and is not the install's current app config.
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	service.InstallOverviewResponse
// @Router					/v1/installs/{install_id}/overview [get]
func (s *service) GetInstallOverview(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	overview, err := s.getInstallOverview(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install overview: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, overview)
}

func (s *service) getInstallOverview(ctx context.Context, orgID, installID string) (*InstallOverviewResponse, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	tracking, err := s.installBranchTracking(ctx, &install)
	if err != nil {
		return nil, err
	}
	drift, err := s.installConfigDrift(ctx, &install)
	if err != nil {
		return nil, err
	}
	return &InstallOverviewResponse{
		BranchTracking: tracking,
		ConfigDrift:    drift,
	}, nil
}

func currentAppConfigID(install *app.Install) string {
	if install.AppConfigRef.ExpectedConfigID != "" {
		return install.AppConfigRef.ExpectedConfigID
	}
	return install.AppConfigID
}

func (s *service) installConfigDrift(ctx context.Context, install *app.Install) (InstallConfigDrift, error) {
	current := currentAppConfigID(install)
	drift := InstallConfigDrift{
		CurrentAppConfigID: current,
		Components:         []InstallConfigDriftComponent{},
	}

	var stack app.InstallStack
	err := s.db.WithContext(ctx).
		Where(app.InstallStack{InstallID: install.ID, OrgID: install.OrgID}).
		First(&stack).Error
	if err == nil {
		drift.Stack = &InstallConfigDriftResource{
			Drifted:            configDrifted(current, stack.AppConfigRef.AppliedConfigID),
			AppliedAppConfigID: stack.AppConfigRef.AppliedConfigID,
		}
	} else if !isNotFound(err) {
		return drift, fmt.Errorf("unable to get install stack: %w", err)
	}

	var sandbox app.InstallSandbox
	err = s.db.WithContext(ctx).
		Where(app.InstallSandbox{InstallID: install.ID, OrgID: install.OrgID}).
		First(&sandbox).Error
	if err == nil {
		drift.Sandbox = &InstallConfigDriftResource{
			Drifted:            configDrifted(current, sandbox.AppConfigRef.AppliedConfigID),
			AppliedAppConfigID: sandbox.AppConfigRef.AppliedConfigID,
		}
	} else if !isNotFound(err) {
		return drift, fmt.Errorf("unable to get install sandbox: %w", err)
	}

	var components []app.InstallComponent
	if err := s.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: install.ID, OrgID: install.OrgID}).
		Find(&components).Error; err != nil {
		return drift, fmt.Errorf("unable to list install components: %w", err)
	}
	for i := range components {
		if components[i].Component.Type.IsImage() {
			continue
		}
		drift.Components = append(drift.Components, InstallConfigDriftComponent{
			Name:               components[i].Component.Name,
			ComponentID:        components[i].ComponentID,
			Drifted:            configDrifted(current, components[i].AppConfigRef.AppliedConfigID),
			AppliedAppConfigID: components[i].AppConfigRef.AppliedConfigID,
		})
	}
	return drift, nil
}

func configDrifted(current, applied string) bool {
	return applied != "" && current != "" && applied != current
}

func (d InstallConfigDrift) driftedCount() int {
	n := 0
	if d.Stack != nil && d.Stack.Drifted {
		n++
	}
	if d.Sandbox != nil && d.Sandbox.Drifted {
		n++
	}
	for i := range d.Components {
		if d.Components[i].Drifted {
			n++
		}
	}
	return n
}

func (s *service) installBranchTracking(ctx context.Context, install *app.Install) (InstallBranchTracking, error) {
	tracking := InstallBranchTracking{Status: "current", PendingCommits: []InstallOverviewCommit{}}

	var connection app.InstallAppBranchConnection
	err := s.db.WithContext(ctx).
		Preload("AppBranch").
		Where(app.InstallAppBranchConnection{InstallID: install.ID, OrgID: install.OrgID, Active: true}).
		Order("created_at DESC, id DESC").
		First(&connection).Error
	if err != nil && !isNotFound(err) {
		return tracking, fmt.Errorf("unable to get install app branch: %w", err)
	}
	if err == nil {
		tracking.BranchID = connection.AppBranchID
		tracking.TargetBranch = connection.AppBranch.Name
		repo, gitBranch, directory, repoErr := s.branchRepo(ctx, connection.AppBranchID)
		if repoErr != nil {
			return tracking, repoErr
		}
		tracking.Repo = repo
		tracking.GitBranch = gitBranch
		tracking.Directory = directory
	}

	expected, err := s.latestBranchRun(ctx, tracking.BranchID)
	if err != nil {
		return tracking, err
	}
	applied, err := s.appliedBranchRun(ctx, install)
	if err != nil {
		return tracking, err
	}
	tracking.ExpectedCommit = commitFromRun(expected)
	tracking.AppliedCommit = commitFromRun(applied)
	tracking.Status = branchTrackingStatus(expected, applied)
	if err := s.installPendingCommits(ctx, install, &tracking); err != nil {
		return tracking, err
	}
	return tracking, nil
}

func installTrackingRuns(db *gorm.DB, orgID, branchID string) *gorm.DB {
	return db.Model(&app.AppBranchRun{}).
		Where(app.AppBranchRun{OrgID: orgID, AppBranchID: branchID}).
		Where("app_branch_runs.no_config_changes = ?", false).
		Where("app_branch_runs.plan_only = ?", false).
		Where("app_branch_runs.run_type IS DISTINCT FROM ?", app.AppBranchRunTypeGitPreview).
		Where("NOT EXISTS (SELECT 1 FROM app_branch_run_previews p WHERE p.app_branch_run_id = app_branch_runs.id AND p.deleted_at = 0)")
}

func (s *service) installPendingCommits(ctx context.Context, install *app.Install, tracking *InstallBranchTracking) error {
	if tracking.BranchID == "" || install.AppConfigID == "" {
		return nil
	}
	var selectedRun app.AppBranchRun
	err := installTrackingRuns(s.db.WithContext(ctx), install.OrgID, tracking.BranchID).
		Where(app.AppBranchRun{AppConfigID: install.AppConfigID}).
		Preload("VCSConnectionCommit").
		Order("created_at DESC, id DESC").
		First(&selectedRun).Error
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to get selected app branch run: %w", err)
	}

	var selectedConfig app.AppConfig
	err = s.db.WithContext(ctx).
		Select("id, created_at").
		Where(app.AppConfig{ID: install.AppConfigID, OrgID: install.OrgID, AppID: install.AppID}).
		First(&selectedConfig).Error
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("unable to get selected app config: %w", err)
	}
	tracking.SelectedCommit = commitFromRun(&selectedRun)

	newerConfigs := s.db.WithContext(ctx).Model(&app.AppConfig{}).
		Where(app.AppConfig{OrgID: install.OrgID, AppID: install.AppID}).
		Where("created_at > ? OR (created_at = ? AND id > ?)", selectedConfig.CreatedAt, selectedConfig.CreatedAt, selectedConfig.ID).
		Where("id IN (?)", installTrackingRuns(s.db.WithContext(ctx), install.OrgID, tracking.BranchID).Select("app_config_id"))
	var count int64
	if err := newerConfigs.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return fmt.Errorf("unable to count newer app configs: %w", err)
	}
	tracking.CommitsBehind = &count
	if count == 0 {
		return nil
	}

	var configIDs []string
	if err := newerConfigs.Order("created_at DESC, id DESC").Limit(50).Pluck("id", &configIDs).Error; err != nil {
		return fmt.Errorf("unable to list newer app configs: %w", err)
	}
	var runs []app.AppBranchRun
	if err := installTrackingRuns(s.db.WithContext(ctx), install.OrgID, tracking.BranchID).
		Select("DISTINCT ON (app_config_id) app_branch_runs.*").
		Where("app_config_id IN ?", configIDs).
		Preload("VCSConnectionCommit").
		Order("app_config_id, created_at DESC, id DESC").
		Find(&runs).Error; err != nil {
		return fmt.Errorf("unable to get pending app branch runs: %w", err)
	}
	workflowIDs := make([]string, 0, len(runs))
	for _, run := range runs {
		if run.WorkflowID != nil {
			workflowIDs = append(workflowIDs, *run.WorkflowID)
		}
	}
	var awaitingWorkflowIDs []string
	if len(workflowIDs) > 0 {
		if err := s.db.WithContext(ctx).Model(&app.WorkflowStep{}).
			Distinct().
			Joins("JOIN install_workflow_step_approvals approvals ON approvals.install_workflow_step_id = install_workflow_steps.id AND approvals.deleted_at = 0").
			Joins("LEFT JOIN install_workflow_step_approval_responses responses ON responses.install_workflow_step_approval_id = approvals.id AND responses.deleted_at = 0").
			Where("install_workflow_steps.install_workflow_id IN ?", workflowIDs).
			Where("install_workflow_steps.execution_type = ?", app.WorkflowStepExecutionTypeApproval).
			Where("install_workflow_steps.status->>'status' = ?", string(app.AwaitingApproval)).
			Where("responses.id IS NULL").
			Pluck("install_workflow_steps.install_workflow_id", &awaitingWorkflowIDs).Error; err != nil {
			return fmt.Errorf("unable to get pending branch approvals: %w", err)
		}
	}
	awaiting := make(map[string]bool, len(awaitingWorkflowIDs))
	for _, id := range awaitingWorkflowIDs {
		awaiting[id] = true
	}
	commits := make(map[string]*InstallOverviewCommit, len(runs))
	for i := range runs {
		commit := commitFromRun(&runs[i])
		commit.AwaitingApproval = awaiting[commit.WorkflowID]
		commits[commit.AppConfigID] = commit
	}
	for _, id := range configIDs {
		if commit := commits[id]; commit != nil {
			tracking.PendingCommits = append(tracking.PendingCommits, *commit)
		}
	}
	return nil
}

func (s *service) branchRepo(ctx context.Context, branchID string) (repo, gitBranch, directory string, err error) {
	if branchID == "" {
		return "", "", "", nil
	}
	var cfg app.AppBranchConfig
	err = s.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		Where(app.AppBranchConfig{AppBranchID: branchID}).
		Order("created_at DESC").
		First(&cfg).Error
	if isNotFound(err) {
		return "", "", "", nil
	}
	if err != nil {
		return "", "", "", fmt.Errorf("unable to get app branch config: %w", err)
	}
	switch {
	case cfg.ConnectedGithubVCSConfig != nil:
		return cfg.ConnectedGithubVCSConfig.Repo, cfg.ConnectedGithubVCSConfig.Branch, cfg.ConnectedGithubVCSConfig.Directory, nil
	case cfg.PublicGitVCSConfig != nil:
		return cfg.PublicGitVCSConfig.Repo, cfg.PublicGitVCSConfig.Branch, cfg.PublicGitVCSConfig.Directory, nil
	default:
		return "", "", "", nil
	}
}

func (s *service) latestBranchRun(ctx context.Context, branchID string) (*app.AppBranchRun, error) {
	if branchID == "" {
		return nil, nil
	}
	var run app.AppBranchRun
	err := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Where(app.AppBranchRun{AppBranchID: branchID}).
		Order("created_at DESC").
		First(&run).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get latest app branch run: %w", err)
	}
	return &run, nil
}

func (s *service) appliedBranchRun(ctx context.Context, install *app.Install) (*app.AppBranchRun, error) {
	appliedID := install.AppConfigRef.AppliedConfigID
	if appliedID == "" {
		return nil, nil
	}
	var version app.InstallAppConfigVersion
	err := s.db.WithContext(ctx).
		Preload("AppBranchRun.VCSConnectionCommit").
		Where(app.InstallAppConfigVersion{InstallID: install.ID, OrgID: install.OrgID, NewAppConfigID: appliedID}).
		Order("created_at DESC").
		First(&version).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get applied app config version: %w", err)
	}
	if version.AppBranchRunID == nil || *version.AppBranchRunID == "" {
		return nil, nil
	}
	return &version.AppBranchRun, nil
}

func commitFromRun(run *app.AppBranchRun) *InstallOverviewCommit {
	if run == nil {
		return nil
	}
	commit := &InstallOverviewCommit{
		RunID:       run.ID,
		BranchID:    run.AppBranchID,
		AppConfigID: run.AppConfigID,
		RunStatus:   run.Status,
		SHA:         run.RunMetadata().HeadSHA,
	}
	if run.WorkflowID != nil {
		commit.WorkflowID = *run.WorkflowID
	}
	if commit.SHA == "" {
		commit.SHA = run.HeadSHA
	}
	created := run.CreatedAt
	commit.CreatedAt = &created
	if run.VCSConnectionCommit != nil {
		commit.Author = run.VCSConnectionCommit.AuthorName
		commit.Message = firstLine(run.VCSConnectionCommit.Message)
		if commit.SHA == "" {
			commit.SHA = run.VCSConnectionCommit.SHA
		}
	}
	return commit
}

func branchTrackingStatus(expected, applied *app.AppBranchRun) string {
	if expected != nil && !branchRunTerminal(expected.Status) {
		return "updating"
	}
	expectedSHA := ""
	appliedSHA := ""
	if expected != nil {
		expectedSHA = commitFromRun(expected).SHA
	}
	if applied != nil {
		appliedSHA = commitFromRun(applied).SHA
	}
	if expectedSHA != "" && appliedSHA != "" && expectedSHA != appliedSHA {
		return "pending"
	}
	return "current"
}

func branchRunTerminal(status string) bool {
	switch status {
	case "", "pending", "in-progress", "queued", "planning", "executing", "syncing":
		return false
	default:
		return true
	}
}

func firstLine(message string) string {
	message = strings.TrimSpace(message)
	if i := strings.IndexByte(message, '\n'); i >= 0 {
		return message[:i]
	}
	return message
}

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
