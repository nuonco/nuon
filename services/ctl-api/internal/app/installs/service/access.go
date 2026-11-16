package service

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const authorizedInstallKey = "nuon.installs.authorized_install"

func (s *service) installResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Resources(s.db, orgID)
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
