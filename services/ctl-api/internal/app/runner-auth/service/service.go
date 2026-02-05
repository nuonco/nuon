package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	V          *validator.Validate
	DB         *gorm.DB `name:"psql"`
	L          *zap.Logger
	AcctClient *account.Client
}

type service struct {
	v          *validator.Validate
	l          *zap.Logger
	db         *gorm.DB
	acctClient *account.Client
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	return &service{
		v:          params.V,
		l:          params.L,
		db:         params.DB,
		acctClient: params.AcctClient,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	auth := api.Group("/v1/runner-auth")
	{
		auth.POST("/aws", s.RunnerAuthAWS)
	}
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}
