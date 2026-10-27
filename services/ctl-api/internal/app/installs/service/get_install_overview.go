package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/alignment"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallOverviewCommit = alignment.InstallOverviewCommit

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
	RunsBehind     *int64                  `json:"runs_behind,omitempty" extensions:"x-nullable"`
	SuccessfulRuns []InstallOverviewCommit `json:"successful_runs"`
	CommitsBehind  *int64                  `json:"commits_behind,omitempty" extensions:"x-nullable,x-deprecated"`
	PendingCommits []InstallOverviewCommit `json:"pending_commits" extensions:"x-deprecated"`
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
	Alignment      *InstallAlignment     `json:"alignment,omitempty"`
}

// @ID						GetInstallOverview
// @Summary				install overview
// @Description			Returns branch tracking, config drift, and per-resource alignment for an install. runs_behind counts successful app branch runs with newer app configurations on the tracked branch relative to the install's selected app_config_id, excluding previews and no-config-change runs. It is omitted when selected branch provenance is unknown. successful_runs includes the newest 50 successful runs with their associated commits, ordered by run creation time. commits_behind and pending_commits are deprecated aliases of runs_behind and successful_runs. config_drift compares applied app config IDs for compatibility. alignment compares resources against the selected app config's builds and relevant inputs, separating changes not applied from state drift detected by scans.
// @Param					install_id	path	string	true	"install ID"
// @Param                 include_alignment query bool false "include per-resource alignment; set false for branch tracking and config drift only" Default(true)
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

	overview, err := s.getInstallOverview(ctx, org.ID, ctx.Param("install_id"), ctx.Query("include_alignment") != "false")
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install overview: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, overview)
}

func (s *service) getInstallOverview(ctx context.Context, orgID, installID string, includeAlignment bool) (*InstallOverviewResponse, error) {
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
	tracking.CommitsBehind = tracking.RunsBehind
	tracking.PendingCommits = tracking.SuccessfulRuns
	overview := &InstallOverviewResponse{
		BranchTracking: tracking,
		ConfigDrift:    drift,
	}
	if !includeAlignment {
		return overview, nil
	}
	alignment, err := s.alignment.Read(ctx, &install)
	if err != nil {
		alignment = &InstallAlignment{
			Error:     "Resource alignment unavailable",
			Resources: []InstallAlignmentResource{},
		}
	} else {
		err = s.alignmentBranchRunCounts(ctx, &install, tracking.BranchID, alignment)
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		s.l.Warn("unable to complete install alignment",
			zap.String("install_id", install.ID),
			zap.String("org_id", install.OrgID),
			zap.Error(err))
	}
	overview.Alignment = alignment
	return overview, nil
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
	tracking := InstallBranchTracking{Status: "current", SuccessfulRuns: []InstallOverviewCommit{}}

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
	if err := s.installSuccessfulBranchRuns(ctx, install, &tracking); err != nil {
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

func (s *service) installSuccessfulBranchRuns(ctx context.Context, install *app.Install, tracking *InstallBranchTracking) error {
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
		Select("id")
	newerRuns := installTrackingRuns(s.db.WithContext(ctx), install.OrgID, tracking.BranchID).
		Where(app.AppBranchRun{Status: "success"}).
		Where("app_config_id IN (?)", newerConfigs)
	var count int64
	if err := newerRuns.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return fmt.Errorf("unable to count successful app branch runs: %w", err)
	}
	tracking.RunsBehind = &count
	if count == 0 {
		return nil
	}

	var runs []app.AppBranchRun
	if err := newerRuns.
		Preload("VCSConnectionCommit").
		Order("created_at DESC, id DESC").
		Limit(50).
		Find(&runs).Error; err != nil {
		return fmt.Errorf("unable to get successful app branch runs: %w", err)
	}
	for i := range runs {
		tracking.SuccessfulRuns = append(tracking.SuccessfulRuns, *commitFromRun(&runs[i]))
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
	return alignment.CommitFromRun(run)
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

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
