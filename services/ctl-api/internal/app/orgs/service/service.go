package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v  *validator.Validate
	l  *zap.Logger
	db *gorm.DB
	mw metrics.Writer
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/v1/org/:id", s.GetOrg)
	api.DELETE("/v1/org/:id", s.DeleteOrg)
	api.PATCH("/v1/org/:id", s.UpdateOrg)
	api.POST("/v1/org", s.CreateOrg)

	api.POST("/v1/org/:org_id/user", s.CreateUser)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/all-orgs", s.GetAllOrgs)
	api.POST("/v1/org/:org_id/support-users", s.CreateSupportUsers)
	return nil
}

func New(v *validator.Validate, db *gorm.DB, mw metrics.Writer, l *zap.Logger) *service {
	return &service{
		l:  l,
		v:  v,
		db: db,
		mw: mw,
	}
}
