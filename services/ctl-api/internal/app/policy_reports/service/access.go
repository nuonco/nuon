package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	appaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/access"
	componentaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/components/access"
	installaccess "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/access"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *service) reportResources(ctx *gin.Context) (*authz.Resources, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	db := s.db.WithContext(ctx.Request.Context()).Session(&gorm.Session{NewDB: true})
	owners := db.Raw(`SELECT 'install_deploys'::text AS owner_type, d.* FROM (?) d
		UNION ALL SELECT 'install_sandbox_runs', r.* FROM (?) r
		UNION ALL SELECT 'component_builds', b.* FROM (?) b`,
		installaccess.Deploys.Query(db, orgID), installaccess.SandboxRuns.Query(db, orgID), componentaccess.Builds.Query(db, orgID))
	query := db.Raw(`SELECT r.id, r.org_id, o.app_id, o.install_id FROM policy_reports r
		JOIN (?) o ON o.id = r.owner_id AND o.owner_type = r.owner_type AND o.org_id = r.org_id
		WHERE r.deleted_at = 0`, owners)
	return authz.NewResources(query, permissions.KindOperation, orgID)
}

func (s *service) reportTarget(ctx *gin.Context) (permissions.Target, error) {
	resources, err := s.reportResources(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), ctx.Param("report_id"))
}

func (s *service) analyticsTarget(ctx *gin.Context) (permissions.Target, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return permissions.Target{}, err
	}
	resources, err := appaccess.Apps.In(s.db, orgID)
	if err != nil {
		return permissions.Target{}, err
	}
	return resources.Resolve(ctx.Request.Context(), ctx.Param("app_id"))
}
