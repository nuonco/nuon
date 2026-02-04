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

// RouterOptions configures the test router setup.
type RouterOptions struct {
	// Logger for middlewares
	L *zap.Logger
	// Database for pagination and patcher middleware
	DB *gorm.DB
	// TestOrg to inject into context (optional)
	TestOrg *app.Org
	// TestAccount to inject into context (optional)
	TestAcc *app.Account
	// AdditionalMiddlewares to add after standard middlewares (optional)
	AdditionalMiddlewares []gin.HandlerFunc
}

// NewTestRouter creates a gin router with standard test middlewares.
//
// Standard middlewares included (in order):
// 1. stderr - Error handling and JSON error responses (REQUIRED)
// 2. patcher - PATCH request field extraction for partial updates
// 3. pagination - Query parameter parsing for paginated endpoints
// 4. context injection - Injects test org and account into context
//
// Usage:
//
//	router := testfx.NewTestRouter(testfx.RouterOptions{
//	    L:       s.service.L,
//	    DB:      s.service.DB,
//	    TestOrg: s.testOrg,
//	    TestAcc: s.testAcc,
//	})
//
// With additional middlewares:
//
//	router := testfx.NewTestRouter(testfx.RouterOptions{
//	    L:       s.service.L,
//	    DB:      s.service.DB,
//	    TestOrg: s.testOrg,
//	    TestAcc: s.testAcc,
//	    AdditionalMiddlewares: []gin.HandlerFunc{
//	        myCustomMiddleware.Handler(),
//	    },
//	})
func NewTestRouter(opts RouterOptions) *gin.Engine {
	router := gin.New()

	// 1. CRITICAL: Add stderr middleware FIRST
	// Without this, error responses will have empty bodies
	errMiddleware := stderr.New(opts.L, nil)
	router.Use(errMiddleware.Handler())

	// 2. Add patcher middleware for PATCH request handling
	patcherMW := patcher.New(patcher.Params{
		L:  opts.L,
		DB: opts.DB,
	})
	router.Use(patcherMW.Handler())

	// 3. Add pagination middleware for paginated GET endpoints
	paginationMW := pagination.New(pagination.Params{
		L:  opts.L,
		DB: opts.DB,
	})
	router.Use(paginationMW.Handler())

	// 4. Add any additional middlewares provided by caller
	for _, mw := range opts.AdditionalMiddlewares {
		router.Use(mw)
	}

	// 5. Add context injection middleware (must be last)
	// This injects test org and account context for authentication
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
