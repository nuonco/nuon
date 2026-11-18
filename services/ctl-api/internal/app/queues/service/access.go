package service

import (
	"github.com/gin-gonic/gin"

	orgaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/queues/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) queueResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return access.Queues.In(s.db, orgID)
}

func (s *service) queueTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.queueResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), ctx.Param("queue_id"))
}

func (s *service) graphTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := orgaccess.Orgs.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), orgID)
}
