package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/hooks"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v     *validator.Validate
	l     *zap.Logger
	db    *gorm.DB
	mw    metrics.Writer
	hooks *hooks.Hooks
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs/:id", s.GetOrg)
	api.DELETE("/v1/orgs/:id", s.DeleteOrg)
	api.PATCH("/v1/orgs/:id", s.UpdateOrg)
	api.POST("/v1/orgs", s.CreateOrg)
	//api.GET("/v1/orgs", s.GetCurrentUserOrgs)

	api.POST("/v1/orgs/:org_id/user", s.CreateUser)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/orgs", s.GetAllOrgs)
	api.POST("/v1/orgs/:org_id/support-users", s.CreateSupportUsers)
	return nil
}

func New(v *validator.Validate, db *gorm.DB, mw metrics.Writer, l *zap.Logger, hooks *hooks.Hooks) *service {
	return &service{
		l:     l,
		v:     v,
		db:    db,
		mw:    mw,
		hooks: hooks,
	}
}
