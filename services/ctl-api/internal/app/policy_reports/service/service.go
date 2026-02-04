package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	V             *validator.Validate
	DB            *gorm.DB `name:"psql"`
	MW            metrics.Writer
	L             *zap.Logger
	Cfg           *internal.Config
	EndpointAudit *api.EndpointAudit
}

type service struct {
	api.RouteRegister
	v   *validator.Validate
	db  *gorm.DB
	mw  metrics.Writer
	l   *zap.Logger
	cfg *internal.Config
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	policyReports := ge.Group("/v1/policy-reports")
	{
		policyReports.GET("", s.GetPolicyReports)
		policyReports.GET("/:report_id", s.GetPolicyReport)
		policyReports.GET("/:report_id/export", s.ExportPolicyReport)
	}
	return nil
}

func (s *service) RegisterInternalRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAuthRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterAdminDashboardRoutes(api *gin.Engine) error {
	return nil
}

func New(params Params) *service {
	return &service{
		RouteRegister: api.RouteRegister{
			EndpointAudit: params.EndpointAudit,
		},
		v:   params.V,
		db:  params.DB,
		mw:  params.MW,
		l:   params.L,
		cfg: params.Cfg,
	}
}
