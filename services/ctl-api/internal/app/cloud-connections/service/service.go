package service

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
	cloudconnectionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections/helpers"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/oidcissuer"
)

type Params struct {
	fx.In

	DB            *gorm.DB `name:"psql"`
	Cfg           *internal.Config
	L             *zap.Logger
	EndpointAudit *apiPkg.EndpointAudit
	Helpers       *cloudconnectionshelpers.Helpers
}

type service struct {
	apiPkg.RouteRegister
	db                  *gorm.DB
	l                   *zap.Logger
	issuer              *oidcissuer.Issuer
	helpers             *cloudconnectionshelpers.Helpers
	enqueueVerification func(context.Context, *app.CloudConnection) error
}

var _ apiPkg.Service = (*service)(nil)
var _ apiPkg.MCPService = (*service)(nil)

func New(params Params) (*service, error) {
	issuer, err := cloudconnections.IssuerFromConfig(params.Cfg)
	if err != nil {
		return nil, err
	}
	return &service{
		RouteRegister:       apiPkg.RouteRegister{EndpointAudit: params.EndpointAudit},
		db:                  params.DB,
		l:                   params.L,
		issuer:              issuer,
		helpers:             params.Helpers,
		enqueueVerification: params.Helpers.EnqueueVerification,
	}, nil
}

func (s *service) RegisterPublicRoutes(api *gin.Engine) error {
	routes := api.Group("/v1/cloud-connections")
	routes.GET("", s.List)
	routes.POST("", s.Create)
	routes.GET("/:connection_id", s.Get)
	routes.DELETE("/:connection_id", s.Delete)
	routes.POST("/:connection_id/verify", s.Verify)
	routes.GET("/:connection_id/setup", s.Setup)
	return nil
}

func (s *service) RegisterAuthRoutes(*gin.Engine) error           { return nil }
func (s *service) RegisterInternalRoutes(*gin.Engine) error       { return nil }
func (s *service) RegisterRunnerRoutes(*gin.Engine) error         { return nil }
func (s *service) RegisterAdminDashboardRoutes(*gin.Engine) error { return nil }
func (s *service) RegisterSlackRoutes(*gin.Engine) error          { return nil }

func (s *service) RegisterMCPTools(server *mcp.Server) {
	mcp.AddTool(server, apiPkg.MCPReadTool("list_cloud_connections", "List cloud connections", "List cloud connections in the current org."+apiPkg.MCPListToolHint), s.mcpList)
	mcp.AddTool(server, apiPkg.MCPReadTool("get_cloud_connection", "Get cloud connection", "Get a cloud connection and its setup material by ID."), s.mcpGet)
	mcp.AddTool(server, apiPkg.MCPWriteTool("create_cloud_connection", "Create cloud connection", "WRITE OPERATION: Create an AWS cloud connection with the stacks or custom preset. For custom, attach your own permissions policy.", false, false), s.mcpCreate)
	mcp.AddTool(server, apiPkg.MCPWriteTool("verify_cloud_connection", "Verify cloud connection", "WRITE OPERATION: Enqueue cloud connection verification. Poll get_cloud_connection until verification_in_progress is false.", false, false), s.mcpVerify)
	mcp.AddTool(server, apiPkg.MCPWriteTool("delete_cloud_connection", "Delete cloud connection", "WRITE OPERATION: Delete an unused cloud connection.", true, true), s.mcpDelete)
}

func (s *service) get(ctx *gin.Context, orgID, id string) (*app.CloudConnection, error) {
	return s.getContext(ctx, orgID, id)
}

func (s *service) getContext(ctx context.Context, orgID, id string) (*app.CloudConnection, error) {
	var connection app.CloudConnection
	result := s.db.WithContext(ctx).Where(app.CloudConnection{OrgID: orgID, ID: id}).First(&connection)
	if result.Error != nil {
		return nil, fmt.Errorf("cloud connection not found: %w", result.Error)
	}
	return &connection, nil
}
