package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/scope"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// appGrantScope narrows an apps list to the apps a deferred grantee is entitled
// to see (granted apps, org, or the parent app of a granted install). No-op
// when the request was authorized org-wide.
func (s *service) appGrantScope(ctx *gin.Context) func(*gorm.DB) *gorm.DB {
	if cctx.OrgAuthorized(ctx) {
		return func(tx *gorm.DB) *gorm.DB { return tx }
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return func(tx *gorm.DB) *gorm.DB { return tx.Where("1 = 0") }
	}

	sets := scope.ForList(acct, permissions.PermissionRead)
	return sets.Apps(s.db, "apps.id", "apps.org_id")
}
