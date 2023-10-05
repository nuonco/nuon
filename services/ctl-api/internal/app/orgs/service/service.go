package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
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
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs", s.GetAllOrgs)
	api.POST("/v1/orgs/:org_id/support-users", s.CreateSupportUsers)
	api.POST("/v1/orgs/:org_id/add-user", s.CreateOrgUser)
	api.POST("/v1/orgs/:org_id/admin-delete", s.AdminDeleteOrg)
	return nil
}

func New(v *validator.Validate,
	db *gorm.DB,
	mw metrics.Writer,
	l *zap.Logger,
	hooks *hooks.Hooks,
	appHooks *apphooks.Hooks,
	installHooks *installhooks.Hooks,
	componentHooks *componenthooks.Hooks,
) *service {
	return &service{
		l:              l,
		v:              v,
		db:             db,
		mw:             mw,
		hooks:          hooks,
		appHooks:       appHooks,
		installHooks:   installHooks,
		componentHooks: componentHooks,
	}
}
