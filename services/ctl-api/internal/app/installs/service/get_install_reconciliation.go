package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const skippedRunsLimit = 50

type InstallReconciliationConfig struct {
	AppConfigID    string                 `json:"app_config_id"`
	Version        int                    `json:"version,omitempty"`
	AppBranchRunID string                 `json:"app_branch_run_id,omitempty"`
	Commit         *InstallOverviewCommit `json:"commit,omitempty"`
}

type InstallReconciliationBlocker struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	WorkflowID string `json:"workflow_id,omitempty"`
}

type InstallReconciliationResponse struct {
	Status             string                         `json:"status" enums:"current,behind,in_progress,blocked"`
	AppBranchID        string                         `json:"app_branch_id,omitempty"`
	AppBranchName      string                         `json:"app_branch_name,omitempty"`
	InstallGroupID     string                         `json:"install_group_id,omitempty"`
	InstallGroupName   string                         `json:"install_group_name,omitempty"`
	Applied            *InstallReconciliationConfig   `json:"applied,omitempty"`
	Target             *InstallReconciliationConfig   `json:"target,omitempty"`
	SkippedRuns        []InstallOverviewCommit        `json:"skipped_runs"`
	SkippedRunCount    int64                          `json:"skipped_run_count"`
	InFlightWorkflowID string                         `json:"in_flight_workflow_id,omitempty"`
	Blockers           []InstallReconciliationBlocker `json:"blockers"`
}

// @ID						GetInstallReconciliation
// @Summary				install reconciliation state
// @Description			Reports whether the install runs its app branch's latest deployable config and what reconciling it would do. The target is the newest completed, non-preview branch run, the same one moving an install to a branch resolves. status is current when the install already runs it, in_progress when a workflow is already rolling it out, blocked when something prevents a reconcile, and behind otherwise. skipped_runs lists the branch runs between the install's applied config and the target. Requires the install-reconciliation org feature.
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
// @Success				200	{object}	service.InstallReconciliationResponse
// @Router					/v1/installs/{install_id}/reconciliation [get]
func (s *service) GetInstallReconciliation(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.requireInstallReconciliation(ctx, org.ID); err != nil {
		ctx.Error(err)
		return
	}

	install, err := s.getReconcilableInstall(ctx, org.ID, ctx.Param("install_id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	target, err := s.helpers.ResolveReconcileTarget(ctx, install)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to resolve reconciliation target: %w", err))
		return
	}
	resp, err := s.installReconciliation(ctx, install, target)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install reconciliation: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func (s *service) requireInstallReconciliation(ctx context.Context, orgID string) error {
	enabled, err := s.featuresClient.OrgHasFeature(ctx, orgID, app.OrgFeatureInstallReconciliation)
	if err != nil {
		return fmt.Errorf("unable to check install reconciliation feature: %w", err)
	}
	if !enabled {
		return stderr.ErrUser{
			Err:         fmt.Errorf("install reconciliation is not enabled for org %s", orgID),
			Description: "Install reconciliation is not enabled for this organization.",
		}
	}
	return nil
}

func (s *service) getReconcilableInstall(ctx context.Context, orgID, installID string) (*app.Install, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).
		Where(app.Install{ID: installID, OrgID: orgID}).
		First(&install).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	return &install, nil
}

func (s *service) installReconciliation(ctx context.Context, install *app.Install, target *installhelpers.ReconcileTarget) (*InstallReconciliationResponse, error) {
	resp := &InstallReconciliationResponse{
		Status:             string(target.Status()),
		AppBranchID:        target.AppBranchID,
		AppBranchName:      target.AppBranchName,
		InstallGroupID:     target.InstallGroupID,
		InstallGroupName:   target.InstallGroupName,
		InFlightWorkflowID: target.InFlightWorkflowID,
		SkippedRuns:        []InstallOverviewCommit{},
		Blockers:           make([]InstallReconciliationBlocker, 0, len(target.Blockers)),
	}
	for _, b := range target.Blockers {
		resp.Blockers = append(resp.Blockers, InstallReconciliationBlocker{
			Code:       string(b.Code),
			Message:    b.Message,
			WorkflowID: b.WorkflowID,
		})
	}
	if !target.HasTarget() {
		return resp, nil
	}

	var tipRun app.AppBranchRun
	if err := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Where(app.AppBranchRun{ID: target.AppBranchRunID, OrgID: install.OrgID}).
		First(&tipRun).Error; err != nil {
		return nil, fmt.Errorf("unable to get target app branch run: %w", err)
	}
	tipConfig, err := s.reconciliationAppConfig(ctx, install, target.AppConfigID)
	if err != nil {
		return nil, err
	}
	resp.Target = &InstallReconciliationConfig{
		AppConfigID:    target.AppConfigID,
		AppBranchRunID: tipRun.ID,
		Commit:         commitFromRun(&tipRun),
	}
	if tipConfig != nil {
		resp.Target.Version = tipConfig.Version
	}

	var appliedConfig *app.AppConfig
	if target.AppliedAppConfigID != "" {
		appliedConfig, err = s.reconciliationAppConfig(ctx, install, target.AppliedAppConfigID)
		if err != nil {
			return nil, err
		}
		resp.Applied = &InstallReconciliationConfig{AppConfigID: target.AppliedAppConfigID}
		if appliedConfig != nil {
			resp.Applied.Version = appliedConfig.Version
		}
		appliedRun, err := s.reconciliationRunForConfig(ctx, install, target.AppBranchID, target.AppliedAppConfigID)
		if err != nil {
			return nil, err
		}
		if appliedRun != nil {
			resp.Applied.AppBranchRunID = appliedRun.ID
			resp.Applied.Commit = commitFromRun(appliedRun)
		}
	}

	if target.Current || tipConfig == nil {
		return resp, nil
	}
	runs, count, err := s.reconciliationSkippedRuns(ctx, install, target.AppBranchID, appliedConfig, tipConfig, tipRun.ID)
	if err != nil {
		return nil, err
	}
	resp.SkippedRuns = runs
	resp.SkippedRunCount = count
	return resp, nil
}

func (s *service) reconciliationAppConfig(ctx context.Context, install *app.Install, configID string) (*app.AppConfig, error) {
	var cfg app.AppConfig
	err := s.db.WithContext(ctx).
		Select("id", "created_at", "version").
		Where(app.AppConfig{ID: configID, OrgID: install.OrgID, AppID: install.AppID}).
		First(&cfg).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get app config %s: %w", configID, err)
	}
	return &cfg, nil
}

func (s *service) reconciliationRunForConfig(ctx context.Context, install *app.Install, branchID, configID string) (*app.AppBranchRun, error) {
	var run app.AppBranchRun
	err := installTrackingRuns(s.db.WithContext(ctx), install.OrgID, branchID).
		Where(app.AppBranchRun{AppConfigID: configID}).
		Preload("VCSConnectionCommit").
		Order("created_at DESC, id DESC").
		First(&run).Error
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get app branch run for config %s: %w", configID, err)
	}
	return &run, nil
}

func (s *service) reconciliationSkippedRuns(ctx context.Context, install *app.Install, branchID string, applied, tip *app.AppConfig, tipRunID string) ([]InstallOverviewCommit, int64, error) {
	configs := s.db.WithContext(ctx).Model(&app.AppConfig{}).
		Where(app.AppConfig{OrgID: install.OrgID, AppID: install.AppID}).
		Where("(created_at, id) <= (?, ?)", tip.CreatedAt, tip.ID).
		Select("id")
	if applied != nil {
		configs = configs.Where("(created_at, id) > (?, ?)", applied.CreatedAt, applied.ID)
	}

	query := installTrackingRuns(s.db.WithContext(ctx), install.OrgID, branchID).
		Where("app_config_id IN (?)", configs).
		Where("app_branch_runs.status = ? OR app_branch_runs.id = ?", "success", tipRunID)

	var count int64
	if err := query.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to count skipped app branch runs: %w", err)
	}

	var runs []app.AppBranchRun
	if err := query.
		Preload("VCSConnectionCommit").
		Order("created_at DESC, id DESC").
		Limit(skippedRunsLimit).
		Find(&runs).Error; err != nil {
		return nil, 0, fmt.Errorf("unable to get skipped app branch runs: %w", err)
	}
	commits := make([]InstallOverviewCommit, 0, len(runs))
	for i := range runs {
		commits = append(commits, *commitFromRun(&runs[i]))
	}
	return commits, count, nil
}
