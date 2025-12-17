package audit

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const auditCacheTTL = 6 * time.Hour

type baseMiddleware struct {
	l             *zap.Logger
	db            *gorm.DB
	context       string
	cfg           *internal.Config
	endpointAudit *api.EndpointAudit
	cache         *sync.Map // stores last write time for each endpoint key
}

func (m baseMiddleware) Name() string {
	return "audit"
}

var skipRoutes = map[string]struct{}{
	"/livez":   {},
	"/readyz":  {},
	"/version": {},
	"/docs/":   {},
}

func (m *baseMiddleware) cacheKey(name, method, route string) string {
	return fmt.Sprintf("%s:%s:%s", name, method, route)
}

func (m *baseMiddleware) shouldWrite(key string) bool {
	now := time.Now()
	if lastWrite, ok := m.cache.Load(key); ok {
		if now.Sub(lastWrite.(time.Time)) < auditCacheTTL {
			return false
		}
	}
	m.cache.Store(key, now)
	return true
}

func (m *baseMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !m.cfg.EnableEndpointAuditing {
			ctx.Next()
			return
		}

		// Skip unmatched routes
		if ctx.FullPath() == "" {
			ctx.Next()
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

		key := m.cacheKey(m.context, ctx.Request.Method, ctx.FullPath())
		if m.shouldWrite(key) {
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

var sharedCache = &sync.Map{}

func newBaseMiddleware(params Params, context string) *baseMiddleware {
	return &baseMiddleware{
		l:             params.L,
		db:            params.DB,
		context:       context,
		cfg:           params.Cfg,
		endpointAudit: params.EndpointAudit,
		cache:         sharedCache,
	}
}
