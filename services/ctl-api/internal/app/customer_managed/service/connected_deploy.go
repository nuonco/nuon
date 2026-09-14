package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
)

func (s *service) PortalDeployRelease(ctx *gin.Context) {
	portal, ok := s.connectedPortal(ctx)
	if !ok {
		return
	}
	if portal.OperatingModel.ApprovalAuthority != app.InstallAuthorityCustomer {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "operating model does not grant customer approval authority"})
		return
	}

	releaseID := ctx.Param("release_id")
	var release app.AppRelease
	if err := s.db.WithContext(ctx).Where(app.AppRelease{
		OrgID: portal.Install.OrgID, AppID: portal.Install.AppID, ID: releaseID, Status: app.AppReleaseStatusReady,
	}).First(&release).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
			return
		}
		ctx.Error(fmt.Errorf("load release: %w", err))
		return
	}

	visible, err := s.releaseVisibleToPortal(ctx, portal.Install, release)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !visible {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "release not found"})
		return
	}

	var active app.InstallReleaseDeployment
	err = s.db.WithContext(ctx).Where(app.InstallReleaseDeployment{
		OrgID: portal.Install.OrgID, InstallID: portal.Install.ID, Status: app.InstallDeploymentStatusSucceeded,
	}).Order("finished_at DESC, created_at DESC, id DESC").First(&active).Error
	if err == nil && active.ReleaseID == release.ID {
		ctx.JSON(http.StatusConflict, gin.H{"error": "release is already active on this install"})
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.Error(fmt.Errorf("load active install release: %w", err))
		return
	}

	var updates []app.InstallAppConfigVersion
	if err := s.db.WithContext(ctx).Preload("Workflow").Where(app.InstallAppConfigVersion{
		OrgID: portal.Install.OrgID, InstallID: portal.Install.ID,
	}).Order("created_at DESC").Find(&updates).Error; err != nil {
		ctx.Error(fmt.Errorf("load release updates: %w", err))
		return
	}
	for _, update := range updates {
		if releaseUpdateActive(update) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "install already has an active release update", "description": "Wait for the current release update to finish before deploying another release."})
			return
		}
	}

	if err := s.validateReleaseBuilds(ctx, release); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update, wf, err := s.installsHelpers.StartInstallAppConfigUpdate(ctx, installshelpers.StartInstallAppConfigUpdateInput{
		InstallID:                portal.Install.ID,
		NewAppConfigID:           release.AppConfigID,
		AppReleaseID:             release.ID,
		OperatingModelID:         portal.OperatingModel.ID,
		ReleaseComponentBuildIDs: release.ComponentBuildIDs,
		ReleaseSandboxBuildID:    release.SandboxBuildID,
		ApprovalOption:           app.InstallApprovalOptionApproveAll,
	})
	if err != nil {
		ctx.Error(pkgerrors.Wrap(err, "start release deployment"))
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"install_app_config_version": update,
		"workflow_id":                wf.ID,
	})
}
