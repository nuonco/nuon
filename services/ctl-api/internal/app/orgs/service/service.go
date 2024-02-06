package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	apphooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/apps/hooks"
	componenthooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/hooks"
	installhooks "github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/hooks"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/hooks"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v              *validator.Validate
	l              *zap.Logger
	db             *gorm.DB
	mw             metrics.Writer
	cfg            *internal.Config
	hooks          *hooks.Hooks
	appHooks       *apphooks.Hooks
	installHooks   *installhooks.Hooks
	componentHooks *componenthooks.Hooks
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	// global routes
	api.POST("/v1/orgs", s.CreateOrg)
	api.GET("/v1/orgs", s.GetCurrentUserOrgs)

	// update your current org
	api.GET("/v1/orgs/current", s.GetOrg)
	api.DELETE("/v1/orgs/current", s.DeleteOrg)
	api.PATCH("/v1/orgs/current", s.UpdateOrg)
	api.POST("/v1/orgs/current/user", s.CreateUser)
	api.GET("/v1/orgs/current/health-checks", s.GetOrgHealthChecks)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs", s.GetAllOrgs)

	api.POST("/v1/orgs/:org_id/admin-add-user", s.CreateOrgUser)
	api.POST("/v1/orgs/:org_id/admin-support-users", s.CreateSupportUsers)
	api.POST("/v1/orgs/:org_id/admin-delete", s.AdminDeleteOrg)
	api.POST("/v1/orgs/:org_id/admin-reprovision", s.AdminReprovisionOrg)
	api.POST("/v1/orgs/:org_id/admin-deprovision", s.AdminDeprovisionOrg)
	api.POST("/v1/orgs/:org_id/admin-restart", s.RestartOrg)
	api.POST("/v1/orgs/:org_id/admin-rename", s.AdminRenameOrg)
	return nil
}

func New(v *validator.Validate,
	db *gorm.DB,
	mw metrics.Writer,
	l *zap.Logger,
	hooks *hooks.Hooks,
	cfg *internal.Config,
	appHooks *apphooks.Hooks,
	installHooks *installhooks.Hooks,
	componentHooks *componenthooks.Hooks,
) *service {
	return &service{
		l:              l,
		v:              v,
		db:             db,
		mw:             mw,
		cfg:            cfg,
		hooks:          hooks,
		appHooks:       appHooks,
		installHooks:   installHooks,
		componentHooks: componentHooks,
	}
}
