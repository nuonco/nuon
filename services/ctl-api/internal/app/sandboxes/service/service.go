package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type service struct {
	v   *validator.Validate
	l   *zap.Logger
	db  *gorm.DB
	mw  metrics.Writer
	cfg *internal.Config
}

func (s *service) RegisterRoutes(api *gin.Engine) error {
	api.GET("/v1/sandboxes", s.GetSandboxes)
	api.GET("/v1/sandboxes/:sandbox_id", s.GetSandbox)
	api.GET("/v1/sandboxes/:sandbox_id/releases", s.GetSandboxReleases)
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	api.POST("/v1/sandboxes", s.CreateSandbox)
	api.POST("/v1/sandboxes/:sandbox_id/release", s.CreateSandboxRelease)
	return nil
}

func New(v *validator.Validate, cfg *internal.Config, db *gorm.DB, mw metrics.Writer, l *zap.Logger) *service {
	return &service{
		cfg: cfg,
		l:   l,
		v:   v,
		db:  db,
		mw:  mw,
	}
}
