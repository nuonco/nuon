package tests

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/pagination"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/patcher"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type RouterOptions struct {
	L                     *zap.Logger
	DB                    *gorm.DB
	TestOrg               *app.Org
	TestAcc               *app.Account
	AdditionalMiddlewares []gin.HandlerFunc
}

func NewTestRouter(opts RouterOptions) *gin.Engine {
	router := gin.New()

	errMiddleware := stderr.New(opts.L, nil)
	router.Use(errMiddleware.Handler())

	patcherMW := patcher.New(patcher.Params{
		L:  opts.L,
		DB: opts.DB,
	})
	router.Use(patcherMW.Handler())

	paginationMW := pagination.New(pagination.Params{
		L:  opts.L,
		DB: opts.DB,
	})
	router.Use(paginationMW.Handler())

	for _, mw := range opts.AdditionalMiddlewares {
		router.Use(mw)
	}

	router.Use(func(c *gin.Context) {
		if opts.TestOrg != nil {
			cctx.SetOrgGinContext(c, opts.TestOrg)
		}
		if opts.TestAcc != nil {
			cctx.SetAccountGinContext(c, opts.TestAcc)
		}
		c.Next()
	})

	return router
}
