package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateReleaseUpdateRequest struct {
	ReleaseID string `json:"release_id" binding:"required"`
}

// @ID CreateInstallReleaseUpdate
// @Summary propose a vendor release to a customer-managed install
// @Tags installs
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param install_id path string true "Install ID"
// @Param req body CreateReleaseUpdateRequest true "Input"
// @Success 201 {object} app.InstallAppConfigVersion
// @Failure 400 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Router /v1/installs/{install_id}/release-updates [post]
func (s *service) CreateReleaseUpdate(ctx *gin.Context) {
	if !s.customerManagedInstallsEnabled(ctx) {
		return
	}
	var req CreateReleaseUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: ctx.Param("install_id"), OrgID: org.ID}).First(&install).Error; err != nil {
		ctx.Error(err)
		return
	}
	var policy app.InstallManagementPolicyVersion
	if err := s.db.WithContext(ctx).Where(app.InstallManagementPolicyVersion{OrgID: org.ID, InstallID: install.ID}).Order("version DESC").First(&policy).Error; err != nil {
		ctx.Error(fmt.Errorf("load install management policy: %w", err))
		return
	}
	if policy.Connectivity != app.InstallConnectivityConnected || policy.ReleaseSelection != app.InstallReleaseSelectionVendor || policy.ApprovalAuthority != app.InstallAuthorityCustomer {
		ctx.Error(stderr.ErrInvalidRequest{Err: fmt.Errorf("install policy does not allow vendor-proposed, customer-approved release updates")})
		return
	}
	var release app.AppRelease
	if err := s.db.WithContext(ctx).Where(app.AppRelease{ID: req.ReleaseID, OrgID: org.ID, AppID: install.AppID, Status: app.AppReleaseStatusReady}).First(&release).Error; err != nil {
		ctx.Error(err)
		return
	}
	var latest app.AppRelease
	if err := s.db.WithContext(ctx).Where(app.AppRelease{
		OrgID: org.ID, AppID: install.AppID, Status: app.AppReleaseStatusReady,
	}).Order("created_at DESC, id DESC").First(&latest).Error; err != nil {
		ctx.Error(err)
		return
	}
	if latest.ID != release.ID {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("only the latest ready release can be proposed")))
		return
	}
	var active app.InstallReleaseDeployment
	err = s.db.WithContext(ctx).Where(app.InstallReleaseDeployment{
		OrgID: org.ID, InstallID: install.ID, Status: app.InstallDeploymentStatusSucceeded,
	}).Order("finished_at DESC, created_at DESC, id DESC").First(&active).Error
	if err == nil && active.ReleaseID == release.ID {
		ctx.Error(stderr.ErrConflict{Err: fmt.Errorf("release is already active on this install")})
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(fmt.Errorf("load active install release: %w", err))
		return
	}
	if err := s.validateReleaseBuilds(ctx, release); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	var updates []app.InstallAppConfigVersion
	if err := s.db.WithContext(ctx).Preload("Workflow").Where(app.InstallAppConfigVersion{OrgID: org.ID, InstallID: install.ID}).Order("created_at DESC").Find(&updates).Error; err != nil {
		ctx.Error(err)
		return
	}
	for _, update := range updates {
		if releaseUpdateActive(update) {
			ctx.Error(stderr.ErrConflict{Err: fmt.Errorf("install already has an active release update"), Description: "Wait for the current release update to finish before proposing another release."})
			return
		}
	}
	update, _, err := s.installsHelpers.StartInstallAppConfigUpdate(ctx, installshelpers.StartInstallAppConfigUpdateInput{
		InstallID: install.ID, NewAppConfigID: release.AppConfigID, AppReleaseID: release.ID,
		PolicyVersionID: policy.ID, ReleaseComponentBuildIDs: release.ComponentBuildIDs,
		ReleaseSandboxBuildID: release.SandboxBuildID,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("start release update: %w", err))
		return
	}
	ctx.JSON(http.StatusCreated, update)
}

func (s *service) validateReleaseBuilds(ctx *gin.Context, release app.AppRelease) error {
	var sandboxCount int64
	if err := s.db.WithContext(ctx).Model(&app.AppSandboxBuild{}).Where(app.AppSandboxBuild{
		ID: release.SandboxBuildID, OrgID: release.OrgID, AppID: release.AppID,
		Status: app.AppSandboxBuildStatusActive,
	}).Count(&sandboxCount).Error; err != nil {
		return err
	}
	if sandboxCount != 1 {
		return fmt.Errorf("release sandbox build is not active or does not belong to the release app")
	}
	cfg, err := s.appsHelpers.GetCustomerManagedAppConfig(ctx, release.OrgID, release.AppID, release.AppConfigID)
	if err != nil {
		return err
	}
	if len(cfg.ComponentConfigConnections) != len(release.ComponentBuildIDs) {
		return fmt.Errorf("release must pin exactly one build for every component config connection")
	}
	for _, connection := range cfg.ComponentConfigConnections {
		buildID := release.ComponentBuildIDs[connection.ID]
		if buildID == "" {
			return fmt.Errorf("release does not pin a build for component config connection %s", connection.ID)
		}
		var build app.ComponentBuild
		if err := s.db.WithContext(ctx).Where(app.ComponentBuild{
			ID: buildID, OrgID: release.OrgID, ComponentConfigConnectionID: connection.ID, Status: app.ComponentBuildStatusActive,
		}).First(&build).Error; err != nil {
			return fmt.Errorf("release component build %s is not active or does not belong to config connection %s: %w", buildID, connection.ID, err)
		}
	}
	return nil
}

func releaseUpdateActive(update app.InstallAppConfigVersion) bool {
	if update.WorkflowID == nil || update.Workflow == nil {
		return false
	}
	status := update.Workflow.Status
	if awaitingRetry, ok := status.Metadata["awaiting_retry"].(bool); status.Status == app.StatusError && ok && awaitingRetry {
		return true
	}
	switch status.Status {
	case app.StatusSuccess, app.StatusError, app.StatusCancelled, app.StatusDiscarded, app.StatusUserSkipped, app.StatusAutoSkipped:
		return false
	default:
		return true
	}
}

func (s *service) AgentListReleaseUpdates(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "release-deployments")
	if !ok {
		return
	}
	var updates []app.InstallAppConfigVersion
	if err := s.db.WithContext(ctx).
		Preload("Workflow").
		Where(app.InstallAppConfigVersion{OrgID: agent.Install.OrgID, InstallID: agent.Install.ID}).
		Where(clause.Neq{Column: "app_release_id", Value: nil}).
		Order("created_at DESC, id DESC").Limit(100).Find(&updates).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, updates)
}
