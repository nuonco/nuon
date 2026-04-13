package service

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	runneraws "github.com/nuonco/nuon/pkg/runner/auth/aws"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type Params struct {
	fx.In

	V          *validator.Validate
	DB         *gorm.DB `name:"psql"`
	L          *zap.Logger
	Cfg        *internal.Config
	AcctClient *account.Client
}

type service struct {
	v          *validator.Validate
	l          *zap.Logger
	db         *gorm.DB
	cfg        *internal.Config
	acctClient *account.Client
	certStore  *runneraws.IIDCertStore
}

var _ api.Service = (*service)(nil)

func New(params Params) *service {
	var certStore *runneraws.IIDCertStore
	if params.L != nil {
		var err error
		certStore, err = runneraws.NewIIDCertStore(params.L, params.Cfg.AWSIIDCertsDir)
		if err != nil {
			params.L.Warn("failed to initialize AWS IID cert store, IID auth will be unavailable", zap.Error(err))
		}
	}

	return &service{
		v:          params.V,
		l:          params.L,
		db:         params.DB,
		cfg:        params.Cfg,
		acctClient: params.AcctClient,
		certStore:  certStore,
	}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	return nil
}

func (s *service) RegisterRunnerRoutes(api *gin.Engine) error {
	auth := api.Group("/v1/runner-auth")
	{
		auth.POST("/aws", s.RunnerAuthAWS)
		auth.POST("/aws-iid", s.RunnerAuthAWSIID)
		auth.POST("/gcp", s.RunnerAuthGCP)
		auth.POST("/azure", s.RunnerAuthAzure)
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
