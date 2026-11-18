package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const authorizedInstallKey = "nuon.installs.authorized_install"

const authorizedLifecycleInstallKey = "nuon.installs.authorized_lifecycle_install"

func (s *service) installLifecycleTarget(ctx *gin.Context) (permissions.Target, error) {
	install, err := s.getInstall(ctx, ctx.Param("install_id"))
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := s.installResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	target, err := resources.Resolve(ctx.Request.Context(), install.ID)
	if err != nil {
		return permissions.Target{}, err
	}
	ctx.Set(authorizedLifecycleInstallKey, install)
	return target, nil
}

func (s *service) installResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Resources(s.db, orgID)
}

func (s *service) installIDTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.installResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, ctx.Param("install_id"))
}

func (s *service) installTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	install, err := s.findInstall(ctx, orgID, ctx.Param("install_id"))
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := s.installResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	target, err := resources.Resolve(ctx, install.ID)
	if err != nil {
		return permissions.Target{}, err
	}
	ctx.Set(authorizedInstallKey, install)
	return target, nil
}

func (s *service) sandboxRunInstallTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var run app.InstallSandboxRun
	if err := s.db.WithContext(ctx.Request.Context()).Select("id", "install_id").
		Where(app.InstallSandboxRun{ID: ctx.Param("run_id"), OrgID: orgID}).Take(&run).Error; err != nil {
		return permissions.Target{}, err
	}
	if id := ctx.Param("install_id"); id != "" && id != run.InstallID {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	resources, err := s.installResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), run.InstallID)
}

func (s *service) stackInstallTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var stack app.InstallStack
	if err := s.db.WithContext(ctx.Request.Context()).Select("id", "install_id").
		Where(app.InstallStack{ID: ctx.Param("stack_id"), OrgID: orgID}).Take(&stack).Error; err != nil {
		return permissions.Target{}, err
	}
	resources, err := s.installResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), stack.InstallID)
}

func (s *service) createInstallTarget(ctx *gin.Context) (permissions.Target, error) {
	appID := ctx.Param("app_id")
	if appID == "" {
		var req CreateInstallV2Request
		if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
			return permissions.Target{}, stderr.NewInvalidRequest(err)
		}
		if err := req.Validate(s.v); err != nil {
			return permissions.Target{}, stderr.NewInvalidRequest(err)
		}
		appID = req.AppID
	}
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := appaccess.Apps.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), appID)
}

func (s *service) installComponentTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.installTarget(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	if target.ID != ctx.Param("install_id") {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	var component app.Component
	if err := s.db.WithContext(ctx.Request.Context()).Select("id").
		Where(app.Component{ID: ctx.Param("component_id"), OrgID: target.OrgID, AppID: target.AppID}).Take(&component).Error; err != nil {
		return permissions.Target{}, err
	}
	var installed app.InstallComponent
	if err := s.db.WithContext(ctx.Request.Context()).Select("id").
		Where(app.InstallComponent{InstallID: target.ID, ComponentID: component.ID, OrgID: target.OrgID}).Take(&installed).Error; err != nil {
		return permissions.Target{}, err
	}
	return target, nil
}

func (s *service) installDeployTarget(ctx *gin.Context) (permissions.Target, error) {
	target, err := s.installIDTarget(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var req CreateInstallDeployRequest
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return permissions.Target{}, stderr.NewInvalidRequest(err)
	}
	if req.BuildID == "" {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	var build app.ComponentBuild
	if err := s.db.WithContext(ctx.Request.Context()).Preload("ComponentConfigConnection").
		Where(app.ComponentBuild{ID: req.BuildID, OrgID: target.OrgID}).Take(&build).Error; err != nil {
		return permissions.Target{}, err
	}
	connection := build.ComponentConfigConnection
	if connection.ID == "" || connection.OrgID != target.OrgID || connection.ComponentID == "" {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	if id := ctx.Param("component_id"); id != "" && id != connection.ComponentID {
		return permissions.Target{}, gorm.ErrRecordNotFound
	}
	var component app.Component
	if err := s.db.WithContext(ctx.Request.Context()).Select("id").
		Where(app.Component{ID: connection.ComponentID, AppID: target.AppID, OrgID: target.OrgID}).Take(&component).Error; err != nil {
		return permissions.Target{}, err
	}
	var installed app.InstallComponent
	if err := s.db.WithContext(ctx.Request.Context()).Select("id").
		Where(app.InstallComponent{InstallID: target.ID, ComponentID: component.ID, OrgID: target.OrgID}).Take(&installed).Error; err != nil {
		return permissions.Target{}, err
	}
	return target, nil
}
