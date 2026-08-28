package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) connectedInstallContext(ctx *gin.Context) {
	orgID := ctx.Request.Header.Get("X-Nuon-Org-ID")
	if orgID == "" {
		orgID = ctx.Query("org_id")
	}
	if orgID == "" {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	cctx.SetOrgGinContext(ctx, &app.Org{ID: orgID})
	ctx.Next()
}

type portalContext struct {
	Install        app.Install
	OperatingModel app.InstallOperatingModel
}

func (s *service) connectedPortal(ctx *gin.Context) (*portalContext, bool) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return nil, false
	}
	enabled, err := s.features.OrgHasFeature(ctx, orgID, app.OrgFeatureCustomerManagedInstalls)
	if err != nil || !enabled {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "customer-managed installs feature is not enabled"})
		return nil, false
	}
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: ctx.Param("install_id"), OrgID: orgID}).First(&install).Error; err != nil {
		ctx.Error(err)
		return nil, false
	}
	var operatingModel app.InstallOperatingModel
	if err := s.db.WithContext(ctx).Where(app.InstallOperatingModel{InstallID: install.ID, OrgID: orgID}).First(&operatingModel).Error; err != nil {
		ctx.Error(err)
		return nil, false
	}
	return &portalContext{Install: install, OperatingModel: operatingModel}, true
}

func (s *service) PortalDiscoverReleases(ctx *gin.Context) {
	portal, ok := s.connectedPortal(ctx)
	if !ok {
		return
	}
	active, err := s.activeInstallRelease(ctx, portal.Install)
	if err != nil {
		ctx.Error(err)
		return
	}
	releaseFilter := s.db.Where(app.AppRelease{AppConfigID: portal.Install.AppConfigID})
	if active.ReleaseID != "" {
		releaseFilter = releaseFilter.Or(app.AppRelease{ID: active.ReleaseID})
	}
	var releases []app.AppRelease
	if err := s.db.WithContext(ctx).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Where(app.AppReleaseMember{OrgID: portal.Install.OrgID}) }).
		Where(app.AppRelease{OrgID: portal.Install.OrgID, AppID: portal.Install.AppID, Status: app.AppReleaseStatusReady}).
		Where(releaseFilter).
		Order("created_at DESC, id DESC").Find(&releases).Error; err != nil {
		ctx.Error(err)
		return
	}
	type discoveredRelease struct {
		Release app.AppRelease `json:"release"`
		Active  bool           `json:"active"`
	}
	result := make([]discoveredRelease, 0, len(releases))
	for _, release := range releases {
		result = append(result, discoveredRelease{Release: release, Active: release.ID == active.ReleaseID})
	}
	ctx.JSON(http.StatusOK, result)
}

func (s *service) PortalGetRelease(ctx *gin.Context) {
	portal, ok := s.connectedPortal(ctx)
	if !ok {
		return
	}
	release, err := s.getRelease(ctx, portal.Install.OrgID, portal.Install.AppID, ctx.Param("release_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	visible, err := s.releaseVisibleToPortal(ctx, portal.Install, *release)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !visible {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, release)
}

func (s *service) PortalGetReleaseFileContent(ctx *gin.Context) {
	portal, ok := s.connectedPortal(ctx)
	if !ok {
		return
	}
	release, err := s.getRelease(ctx, portal.Install.OrgID, portal.Install.AppID, ctx.Param("release_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	visible, err := s.releaseVisibleToPortal(ctx, portal.Install, *release)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !visible {
		ctx.Status(http.StatusNotFound)
		return
	}
	content, err := s.getReleaseFileContent(ctx, portal.Install.OrgID, portal.Install.AppID, release.ID, ctx.Query("path"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, content)
}

func (s *service) PortalGetReleasePackage(ctx *gin.Context) {
	portal, ok := s.connectedPortal(ctx)
	if !ok {
		return
	}
	var pkg app.ReleasePackage
	if err := s.db.WithContext(ctx).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackageMember{OrgID: portal.Install.OrgID}) }).
		Joins("JOIN app_releases ON app_releases.id = release_packages.release_id").
		Where(app.ReleasePackage{ID: ctx.Param("package_id"), OrgID: portal.Install.OrgID}).
		Where("app_releases.app_id = ?", portal.Install.AppID).
		Where("app_releases.status = ?", app.AppReleaseStatusReady).
		First(&pkg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Status(http.StatusNotFound)
			return
		}
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, pkg)
}

func (s *service) activeInstallRelease(ctx context.Context, install app.Install) (app.InstallReleaseDeployment, error) {
	var active app.InstallReleaseDeployment
	err := s.db.WithContext(ctx).
		Where(app.InstallReleaseDeployment{OrgID: install.OrgID, InstallID: install.ID, Status: app.InstallDeploymentStatusSucceeded}).
		Order("finished_at DESC, created_at DESC, id DESC").
		First(&active).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return app.InstallReleaseDeployment{}, nil
	}
	return active, err
}

func (s *service) releaseVisibleToPortal(ctx context.Context, install app.Install, release app.AppRelease) (bool, error) {
	if release.AppConfigID == install.AppConfigID {
		return true, nil
	}
	active, err := s.activeInstallRelease(ctx, install)
	if err != nil {
		return false, err
	}
	return release.ID == active.ReleaseID, nil
}
