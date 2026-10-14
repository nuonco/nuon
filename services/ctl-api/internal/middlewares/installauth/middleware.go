package installauth

import (
	"errors"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const installPrefix = "inl"

type Middleware struct {
	l    *zap.Logger
	db   *gorm.DB
	mu   sync.RWMutex
	orgs map[string]string
}

var _ middlewares.Middleware = (*Middleware)(nil)

type Params struct {
	fx.In

	L  *zap.Logger
	DB *gorm.DB `name:"psql"`
}

func New(params Params) *Middleware {
	return &Middleware{
		l:    params.L,
		db:   params.DB,
		orgs: map[string]string{},
	}
}

func (m *Middleware) Name() string {
	return "install"
}

func (m *Middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsGlobal(ctx) || cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		installID := installIDFromRequest(ctx)
		if installID == "" {
			ctx.Next()
			return
		}

		org, err := cctx.OrgFromContext(ctx)
		if err != nil {
			ctx.Error(stderr.ErrAuthorization{
				Err:         err,
				Description: "no org was set in the middleware",
			})
			ctx.Abort()
			return
		}

		ownerOrgID, err := m.ownerOrgID(ctx, installID)
		if err != nil || ownerOrgID != org.ID {
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				ctx.Error(stderr.ErrSystem{
					Err:         err,
					Description: "unable to authorize install",
				})
				ctx.Abort()
				return
			}
			m.l.Info("install access denied",
				zap.String("install_id", installID),
				zap.String("org_id", org.ID),
				zap.Error(err),
			)
			ctx.Error(stderr.ErrNotFound{
				Err:         errors.New("install not found"),
				Description: "install not found",
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func (m *Middleware) ownerOrgID(ctx *gin.Context, installID string) (string, error) {
	if orgID, ok := m.cachedOrgID(installID); ok {
		return orgID, nil
	}

	var install app.Install
	res := m.db.WithContext(ctx.Request.Context()).
		Select("org_id").
		Where(&app.Install{ID: installID}).
		Take(&install)
	if res.Error != nil {
		return "", res.Error
	}

	m.remember(installID, install.OrgID)
	return install.OrgID, nil
}

func (m *Middleware) cachedOrgID(installID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	orgID, ok := m.orgs[installID]
	return orgID, ok
}

func (m *Middleware) remember(installID, orgID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.orgs[installID] = orgID
}

func installIDFromRequest(ctx *gin.Context) string {
	if id := ctx.Param("install_id"); isInstallID(id) {
		return id
	}
	for _, part := range strings.Split(ctx.Request.URL.Path, "/") {
		if isInstallID(part) {
			return part
		}
	}
	return ""
}

func isInstallID(id string) bool {
	prefix, err := shortid.GetPrefix(id)
	return err == nil && prefix == installPrefix
}
