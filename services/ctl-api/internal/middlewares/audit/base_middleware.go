package audit

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/api"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type baseMiddleware struct {
	l             *zap.Logger
	db            *gorm.DB
	context       string
	cfg           *internal.Config
	endpointAudit *api.EndpointAudit
}

func (m baseMiddleware) Name() string {
	return "audit"
}

var skipRoutes = map[string]struct{}{
	"/livez":   {},
	"/version": {},
	"/docs/":   {},
}

func (m *baseMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !m.cfg.EnableEndpointAuditing {
			return
		}

		if _, ok := skipRoutes[ctx.FullPath()]; ok {
			ctx.Next()
			return
		}

		for route := range skipRoutes {
			if strings.HasPrefix(ctx.FullPath(), route) {
				ctx.Next()
				return
			}
		}

		deprecated := m.endpointAudit.IsDeprecated(
			ctx.Request.Method,
			m.context,
			ctx.FullPath(),
		)
		ea := app.EndpointAudit{
			Method:     ctx.Request.Method,
			Name:       m.context,
			Route:      ctx.FullPath(),
			LastUsedAt: generics.NewNullTime(time.Now()),
			Deprecated: deprecated,
		}
		if res := m.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "deleted_at"},
					{Name: "method"},
					{Name: "name"},
					{Name: "route"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"last_used_at",
					"deprecated",
				}),
			}).
			Create(&ea); res.Error != nil {
			ctx.Error(stderr.ErrSystem{
				Err:         errors.Wrap(res.Error, "unable to emit api endpoint audit"),
				Description: "unable to emit api endpoint auditing",
			})
		}
		if deprecated {
			metricsCtx, err := cctx.MetricsContextFromGinContext(ctx)
			if err != nil {
				m.l.Error("no metrics context found")
				return
			}
			metricsCtx.IsDeprecated = true
		}
	}
}

type Params struct {
	fx.In
	L             *zap.Logger
	DB            *gorm.DB `name:"psql"`
	Cfg           *internal.Config
	EndpointAudit *api.EndpointAudit
}

func newBaseMiddleware(params Params, context string) *baseMiddleware {
	return &baseMiddleware{
		l:             params.L,
		db:            params.DB,
		context:       context,
		cfg:           params.Cfg,
		endpointAudit: params.EndpointAudit,
	}
}
