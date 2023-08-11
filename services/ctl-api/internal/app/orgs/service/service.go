package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"gorm.io/gorm"
)

type service struct {
	v  *validator.Validate
	db *gorm.DB
	mw metrics.Writer
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/v1/org/:org_id", s.GetOrg)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.GET("/v1/all-orgs", s.GetAllOrgs)
	return nil
}

func New(v *validator.Validate, db *gorm.DB, mw metrics.Writer) *service {
	return &service{
		v:  v,
		db: db,
		mw: mw,
	}
}
