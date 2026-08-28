package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type agentContext struct {
	Install app.Install
	Policy  app.InstallManagementPolicyVersion
}

func (s *service) connectedAgent(ctx *gin.Context, _ string) (*agentContext, bool) {
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
	var policy app.InstallManagementPolicyVersion
	if err := s.db.WithContext(ctx).Where(app.InstallManagementPolicyVersion{InstallID: install.ID, OrgID: orgID}).Order("version DESC").First(&policy).Error; err != nil {
		ctx.Error(err)
		return nil, false
	}
	return &agentContext{Install: install, Policy: policy}, true
}

func (s *service) AgentDiscoverReleases(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "")
	if !ok {
		return
	}
	var releases []app.AppRelease
	if err := s.db.WithContext(ctx).
		Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Where(app.AppReleaseMember{OrgID: agent.Install.OrgID}) }).
		Preload("Packages", func(db *gorm.DB) *gorm.DB { return db.Where(app.ReleasePackage{OrgID: agent.Install.OrgID}) }).
		Where(app.AppRelease{OrgID: agent.Install.OrgID, AppID: agent.Install.AppID, AppConfigID: agent.Install.AppConfigID, Status: app.AppReleaseStatusReady}).
		Order("created_at DESC, id DESC").Find(&releases).Error; err != nil {
		ctx.Error(err)
		return
	}
	type discoveredRelease struct {
		Release app.AppRelease `json:"release"`
		Active  bool           `json:"active"`
	}
	result := make([]discoveredRelease, 0, len(releases))
	var active app.InstallReleaseDeployment
	activeErr := s.db.WithContext(ctx).Where(app.InstallReleaseDeployment{OrgID: agent.Install.OrgID, InstallID: agent.Install.ID, Status: app.InstallDeploymentStatusSucceeded}).Order("finished_at DESC, created_at DESC, id DESC").First(&active).Error
	if activeErr != nil && !errors.Is(activeErr, gorm.ErrRecordNotFound) {
		ctx.Error(activeErr)
		return
	}
	for _, release := range releases {
		result = append(result, discoveredRelease{Release: release, Active: release.ID == active.ReleaseID})
	}
	ctx.JSON(http.StatusOK, result)
}

func (s *service) AgentGetRelease(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "")
	if !ok {
		return
	}
	release, err := s.getRelease(ctx, agent.Install.OrgID, agent.Install.AppID, ctx.Param("release_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	if release.AppConfigID != agent.Install.AppConfigID {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, release)
}

func (s *service) AgentGetReleaseFileContent(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "")
	if !ok {
		return
	}
	release, err := s.getRelease(ctx, agent.Install.OrgID, agent.Install.AppID, ctx.Param("release_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	if release.AppConfigID != agent.Install.AppConfigID {
		ctx.Status(http.StatusNotFound)
		return
	}
	content, err := s.getReleaseFileContent(ctx, agent.Install.OrgID, agent.Install.AppID, release.ID, ctx.Query("path"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, content)
}

func (s *service) AgentGetReleasePackage(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "")
	if !ok {
		return
	}
	pkg, err := s.getReleasePackage(ctx, agent.Install.OrgID, ctx.Param("package_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	if pkg.Release.AppID != agent.Install.AppID || pkg.Release.AppConfigID != agent.Install.AppConfigID {
		ctx.Status(http.StatusNotFound)
		return
	}
	ctx.JSON(http.StatusOK, pkg)
}

func (s *service) AgentCreatePackageGrant(ctx *gin.Context) {
	agent, ok := s.connectedAgent(ctx, "")
	if !ok {
		return
	}
	var pkg app.ReleasePackage
	if err := s.db.WithContext(ctx).Preload("Release").Where(app.ReleasePackage{ID: ctx.Param("package_id"), OrgID: agent.Install.OrgID, Status: app.ReleasePackageStatusActive}).First(&pkg).Error; err != nil {
		ctx.Error(err)
		return
	}
	if pkg.Release.AppID != agent.Install.AppID || pkg.Release.AppConfigID != agent.Install.AppConfigID {
		ctx.Status(http.StatusNotFound)
		return
	}
	grant, err := s.createReleasePackageDownloadGrant(ctx, agent.Install.OrgID, pkg.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Header("Cache-Control", "no-store")
	ctx.JSON(http.StatusOK, grant)
}
