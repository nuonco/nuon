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
	SHA       string     `json:"sha,omitempty"`
	RunID     string     `json:"run_id,omitempty"`
	Message   string     `json:"message,omitempty"`
	Author    string     `json:"author,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	RunStatus string     `json:"run_status,omitempty"`
}

type InstallBranchTracking struct {
	BranchID       string                 `json:"branch_id,omitempty"`
	TargetBranch   string                 `json:"target_branch,omitempty"`
	Repo           string                 `json:"repo,omitempty"`
	GitBranch      string                 `json:"git_branch,omitempty"`
	Directory      string                 `json:"directory,omitempty"`
	Status         string                 `json:"status"`
	ExpectedCommit *InstallOverviewCommit `json:"expected_commit,omitempty"`
	AppliedCommit  *InstallOverviewCommit `json:"applied_commit,omitempty"`
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
// @Description			Returns branch tracking and config drift for an install. A stack, sandbox, or component is drifted when its applied app config is set and is not the install's current app config.
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
	tracking := InstallBranchTracking{Status: "current"}

	var connection app.InstallAppBranchConnection
	err := s.db.WithContext(ctx).
		Preload("AppBranch").
		Where(app.InstallAppBranchConnection{InstallID: install.ID, OrgID: install.OrgID}).
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
	return tracking, nil
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
		RunID:     run.ID,
		RunStatus: run.Status,
		SHA:       run.RunMetadata().HeadSHA,
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
