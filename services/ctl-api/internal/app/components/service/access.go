package service

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) componentResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Components.In(s.db, orgID)
}

func (s *service) buildResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Builds.In(s.db, orgID)
}

func (s *service) componentTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.componentResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, ctx.Param("component_id"))
}

func (s *service) namedComponentTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	var component *app.Component
	if ctx.Param("app_id") != "" {
		ref := ctx.Param("component_id")
		if ref == "" {
			ref = ctx.Param("component_name_or_id")
		}
		component, err = s.getAppComponent(ctx, ctx.Param("app_id"), ref)
	} else {
		component, err = s.findComponent(ctx, orgID, ctx.Param("component_id"))
	}
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := s.componentResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, component.ID)
}

func (s *service) buildTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.buildResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, ctx.Param("build_id"))
}

func (s *service) componentAppTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := appaccess.Apps.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, ctx.Param("app_id"))
}

func (s *service) componentGraphTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	component, err := s.findComponent(ctx, orgID, ctx.Param("component_id"))
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := appaccess.Apps.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx, component.AppID)
}

func (s *service) componentConfigRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	target, err := s.componentTarget(ctx)
	if err != nil {
		return nil, err
	}
	var req struct {
		AppConfigID string `json:"app_config_id"`
	}
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return nil, stderr.NewInvalidRequest(err)
	}
	if req.AppConfigID != "" {
		var config app.AppConfig
		if err := s.db.WithContext(ctx).Where(app.AppConfig{ID: req.AppConfigID, AppID: target.AppID, OrgID: target.OrgID}).First(&config).Error; err != nil {
			return nil, err
		}
	}
	return []authz.Requirement{{Operation: permissions.ComponentConfigWrite, Target: target}, {Operation: permissions.ComponentBuildCreate, Target: target}}, nil
}

func (s *service) buildAllRequirements(ctx *gin.Context) ([]authz.Requirement, error) {
	target, err := s.componentAppTarget(ctx)
	if err != nil {
		return nil, err
	}
	var components []*app.Component
	if err := s.db.WithContext(ctx).Where(app.Component{AppID: target.ID, OrgID: target.OrgID}).Find(&components).Error; err != nil {
		return nil, err
	}
	resources, err := s.componentResources(ctx)
	if err != nil {
		return nil, err
	}
	requirements := []authz.Requirement{{Operation: permissions.AppComponentsBuild, Target: target}}
	for _, component := range components {
		child, err := resources.Resolve(ctx, component.ID)
		if err != nil {
			return nil, err
		}
		requirements = append(requirements, authz.Requirement{Operation: permissions.ComponentBuildCreate, Target: child})
	}
	ctx.Set("nuon.components.build_all", components)
	return requirements, nil
}
