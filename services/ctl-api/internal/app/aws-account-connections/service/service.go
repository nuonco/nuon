package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type Params struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	Cfg            *internal.Config
	L              *zap.Logger
	EndpointAudit  *apiPkg.EndpointAudit
	FeaturesClient *features.Features
	Verifier       Verifier `optional:"true"`
}

type service struct {
	apiPkg.RouteRegister
	db       *gorm.DB
	cfg      *internal.Config
	l        *zap.Logger
	features *features.Features
	verifier Verifier
}

var _ apiPkg.Service = (*service)(nil)

func New(params Params) *service {
	verifier := params.Verifier
	if verifier == nil {
		verifier = NewAWSVerifier()
	}
	logger := params.L
	if logger == nil {
		logger = zap.NewNop()
	}
	return &service{RouteRegister: apiPkg.RouteRegister{EndpointAudit: params.EndpointAudit}, db: params.DB, cfg: params.Cfg, l: logger, features: params.FeaturesClient, verifier: verifier}
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	routes := api.Group("/v1/aws-account-connections")
	routes.POST("", s.Create)
	routes.GET("", s.List)
	routes.GET("/:connection_id", s.Get)
	routes.PATCH("/:connection_id", s.Patch)
	routes.DELETE("/:connection_id", s.Delete)
	routes.POST("/:connection_id/verify", s.Verify)
	return nil
}

func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }

func (s *service) gate(ctx *gin.Context) (*app.Org, error) {
	enabled, err := s.features.FeatureEnabled(ctx, app.OrgFeatureAWSAccountConnections)
	if err != nil || !enabled {
		return nil, stderr.ErrAuthorization{
			Err:         fmt.Errorf("aws account connections feature is not enabled"),
			Description: "AWS account connections are not enabled for this organization",
		}
	}
	return cctx.OrgFromContext(ctx)
}

func (s *service) get(ctx *gin.Context, orgID, id string) (*app.AWSAccountConnection, error) {
	var connection app.AWSAccountConnection
	result := s.db.WithContext(ctx).Where(app.AWSAccountConnection{OrgID: orgID, ID: id}).First(&connection)
	if result.Error != nil {
		return nil, fmt.Errorf("aws account connection not found: %w", result.Error)
	}
	return &connection, nil
}
