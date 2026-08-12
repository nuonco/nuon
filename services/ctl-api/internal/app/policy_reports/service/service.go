package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type Params struct {
	fx.In

	V             *validator.Validate
	DB            *gorm.DB `name:"psql"`
	CHDB          *gorm.DB `name:"ch"`
	MW            metrics.Writer
	L             *zap.Logger
	Cfg           *internal.Config
	EndpointAudit *api.EndpointAudit
}

type service struct {
	api.RouteRegister
	v    *validator.Validate
	db   *gorm.DB
	chDB *gorm.DB
	mw   metrics.Writer
	l    *zap.Logger
	cfg  *internal.Config
}

var _ api.Service = (*service)(nil)

func (s *service) RegisterPublicRoutes(ge *gin.Engine) error {
	policyReports := ge.Group("/v1/policy-reports")
	{
		policyReports.GET("", s.GetPolicyReports)
		policyReports.GET("/:report_id", s.GetPolicyReport)
		policyReports.GET("/:report_id/export", s.ExportPolicyReport)
	}

	analytics := ge.Group("/v1/apps/:app_id/policy-analytics")
	{
		analytics.GET("/summary", s.GetPolicyAnalyticsSummary)
		analytics.GET("/timeseries", s.GetPolicyAnalyticsTimeseries)
		analytics.GET("/breakdown", s.GetPolicyAnalyticsBreakdown)
	}

	// install-nested, ancestor-scoped reports (bare routes above stay org-tier).
	reports := ge.Group("/v1/installs/:install_id/policy-reports/:report_id")
	reports.Use(s.requireReportInInstall)
	reports.GET("", s.GetInstallPolicyReport)
	reports.GET("/export", s.ExportInstallPolicyReport)

	return nil
}

// requireReportInInstall gates the install-nested policy-report routes: the
// report named by :report_id must carry install_id = :install_id (the
// denormalized owner column), scoping the URL's claim before the reused
// handler runs.
func (s *service) requireReportInInstall(ctx *gin.Context) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get org from context"))
		ctx.Abort()
		return
	}

	installID := ctx.Param("install_id")
	var count int64
	res := s.db.WithContext(ctx).Model(&app.PolicyReport{}).
		Where(app.PolicyReport{ID: ctx.Param("report_id"), OrgID: orgID, InstallID: &installID}).
		Count(&count)
	if res.Error != nil {
		ctx.Error(errors.Wrap(res.Error, "unable to resolve policy report"))
		ctx.Abort()
		return
	}
	if count == 0 {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "policy report not found"})
		return
	}

	ctx.Next()
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
		v:    params.V,
		db:   params.DB,
		chDB: params.CHDB,
		mw:   params.MW,
		l:    params.L,
		cfg:  params.Cfg,
	}
}

func (s *service) RegisterSlackRoutes(api *gin.Engine) error {
	return nil
}
