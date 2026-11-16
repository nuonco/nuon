package service

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const authorizedAppKey = "nuon.apps.authorized_app"

func (s *service) appResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return authz.NewResources(access.Apps(s.db, orgID), permissions.KindApp, orgID)
}

func (s *service) appTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	a, err := s.findApp(ctx, orgID, ctx.Param("app_id"))
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := s.appResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	target, err := resources.Resolve(ctx, a.ID)
	if err != nil {
		return permissions.Target{}, err
	}
	ctx.Set(authorizedAppKey, a)
	return target, nil
}
