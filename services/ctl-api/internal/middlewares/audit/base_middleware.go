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
)

type baseMiddleware struct {
	l       *zap.Logger
	db      *gorm.DB
	context string
	cfg     *internal.Config
}

func (m baseMiddleware) Name() string {
	return "audit"
}

var routeBlacklist = map[string]struct{}{
	"/livez":   {},
	"/version": {},
	"/docs/":   {},
}

func (m *baseMiddleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !m.cfg.EnableEndpointAuditing {
			return
		}

		// skip if in blacklist
		if _, ok := routeBlacklist[ctx.FullPath()]; ok {
			ctx.Next()
			return
		}

		// or route starts with any values in blacklist
		for route := range routeBlacklist {
			if strings.HasPrefix(ctx.FullPath(), route) {
				ctx.Next()
				return
			}
		}

		ea := app.EndpointAudit{
			Method:     ctx.Request.Method,
			Name:       m.context,
			Route:      ctx.FullPath(),
			LastUsedAt: generics.NewNullTime(time.Now()),
		}
		if res := m.db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "deleted_at"},
					{Name: "method"},
					{Name: "name"},
					{Name: "route"},
				},
				UpdateAll: true,
			}).
			Create(&ea); res.Error != nil {
			ctx.Error(stderr.ErrSystem{
				Err:         errors.Wrap(res.Error, "unable to emit api endpoint audit"),
				Description: "unable to emit api endpoint auditing",
			})
		}
	}
}

type Params struct {
	fx.In
	L  *zap.Logger
	DB *gorm.DB `name:"psql"`
}

func New(l *zap.Logger, db *gorm.DB, contex string) *baseMiddleware {
	return &baseMiddleware{
		l:       l,
		db:      db,
		context: contex,
	}
}
