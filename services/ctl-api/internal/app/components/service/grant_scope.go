package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/scope"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// componentGrantScope narrows an org-wide components list to those a deferred
// grantee can see via a parent-app or org grant. Component is not yet a grant
// target, so there is no component-id tier. No-op when authorized org-wide.
func (s *service) componentGrantScope(ctx *gin.Context) func(*gorm.DB) *gorm.DB {
	if cctx.OrgAuthorized(ctx) {
		return func(tx *gorm.DB) *gorm.DB { return tx }
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return func(tx *gorm.DB) *gorm.DB { return tx.Where("1 = 0") }
	}

	sets := scope.ForList(acct, permissions.PermissionRead)
	return sets.Components(s.db, "components.app_id", "apps.org_id")
}
