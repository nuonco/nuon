package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/scope"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

// installGrantScope narrows an installs list to the rows a deferred grantee is
// entitled to see. It is a no-op when the request was authorized org-wide
// (OrgAuthorized), which is always the case with resource grants disabled.
func (s *service) installGrantScope(ctx *gin.Context) func(*gorm.DB) *gorm.DB {
	if cctx.OrgAuthorized(ctx) {
		return func(tx *gorm.DB) *gorm.DB { return tx }
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		return func(tx *gorm.DB) *gorm.DB { return tx.Where("1 = 0") }
	}

	sets := scope.ForList(acct, permissions.PermissionRead)
	return sets.Installs(
		s.db,
		views.TableOrViewName(s.db, &app.Install{}, ".id"),
		views.TableOrViewName(s.db, &app.Install{}, ".app_id"),
		views.TableOrViewName(s.db, &app.Install{}, ".org_id"),
	)
}
