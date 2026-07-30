package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	Cfg    *internal.Config
	V      *validator.Validate
	Logger *zap.Logger
}

type service struct {
	db  *gorm.DB
	cfg *internal.Config
	v   *validator.Validate
	l   *zap.Logger
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		db:  params.DB,
		cfg: params.Cfg,
		v:   params.V,
		l:   params.Logger,
	}
}

func (s *service) RegisterPublicRoutes(engine *gin.Engine) error {
	grants := engine.Group("/v1/grants")
	{
		grants.POST("", s.CreateGrant)
		grants.GET("", s.ListGrants)
		grants.DELETE("/:grant_id", s.DeleteGrant)
	}

	return nil
}

func (s *service) RegisterRunnerRoutes(_ *gin.Engine) error         { return nil }
func (s *service) RegisterAuthRoutes(_ *gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(_ *gin.Engine) error       { return nil }
func (s *service) RegisterAdminDashboardRoutes(_ *gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(_ *gin.Engine) error          { return nil }
