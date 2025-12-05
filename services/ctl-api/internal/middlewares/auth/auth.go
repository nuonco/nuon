package auth

import (
	"fmt"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	accountshelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/accounts/helpers"
	runnershelpers "github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	L               *zap.Logger
	Cfg             *internal.Config
	DB              *gorm.DB `name:"psql"`
	AuthzClient     *authz.Client
	AcctClient      *account.Client
	AccountsHelpers *accountshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
	EvClient        eventloop.Client
}

type middleware struct {
	cfg             *internal.Config
	l               *zap.Logger
	db              *gorm.DB
	authzClient     *authz.Client
	acctClient      *account.Client
	accountsHelpers *accountshelpers.Helpers
	runnersHelpers  *runnershelpers.Helpers
	evClient        eventloop.Client
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cctx.IsPublic(ctx) {
			ctx.Next()
			return
		}

		token, err := jwtmiddleware.AuthHeaderTokenExtractor(ctx.Request)
		if err != nil {
			ctx.Error(stderr.ErrAuthentication{
				Err:         err,
				Description: "Please make sure you set the -H Authorization:Bearer token header",
			})
			ctx.Abort()
			return
		}

		// we extract the token from query params if it was not provided in the header
		qtoken := ctx.Query("token")
		if token == "" && qtoken != "" {
			token = qtoken
		}

		if token == "" {
			ctx.Error(stderr.ErrAuthentication{
				Err:         fmt.Errorf("auth token was empty"),
				Description: "Please make sure you set the -H Authorization:Bearer <token> header or token query param",
			})
			ctx.Abort()

			return
		}

		acctToken, err := m.fetchAccountToken(ctx, token)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		if acctToken != nil {
			acct, err := m.acctClient.FetchAccount(ctx, acctToken.AccountID)
			if err != nil {
				ctx.Error(err)
				ctx.Abort()
				return
			}

			// Detect CLI usage and update journey step
			m.detectCLIUsage(ctx, acct)

			cctx.SetAccountGinContext(ctx, acct)
			ctx.Next()
			return
		}

		// new token, so validate the token
		claims, err := m.validateToken(ctx, token)
		if err != nil {
			ctx.Error(stderr.ErrAuthentication{
				Err:         err,
				Description: "Please make sure the token is valid, and not expired.",
			})
			ctx.Abort()
			return
		}

		// store the token
		acctToken, err = m.saveAccountToken(ctx, token, claims)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to save account token: %w", err))
			ctx.Abort()
			return
		}

		acct, err := m.acctClient.FetchAccount(ctx, acctToken.AccountID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to fetch: %w", err))
			ctx.Abort()
			return
		}

		// Detect CLI usage and update journey step
		m.detectCLIUsage(ctx, acct)

		cctx.SetAccountGinContext(ctx, acct)
		ctx.Next()
	}
}

func (m *middleware) Name() string {
	return "auth"
}

func New(params Params) *middleware {
	return &middleware{
		l:               params.L,
		cfg:             params.Cfg,
		db:              params.DB,
		authzClient:     params.AuthzClient,
		acctClient:      params.AcctClient,
		accountsHelpers: params.AccountsHelpers,
		runnersHelpers:  params.RunnersHelpers,
		evClient:        params.EvClient,
	}
}

// detectCLIUsage detects if a request is coming from the Nuon CLI and updates the journey step
func (m *middleware) detectCLIUsage(ctx *gin.Context, acct *app.Account) {
	userAgent := ctx.GetHeader("User-Agent")

	// Check if the User-Agent indicates CLI usage
	// The Nuon CLI should set a User-Agent like "nuon-cli/v1.2.3" or "nuon/1.2.3"
	if isCLIUserAgent(userAgent) {
		// Update the cli_installed journey step
		if err := m.accountsHelpers.UpdateUserJourneyStepForCLIInstalled(ctx, acct.ID); err != nil {
			// Log but don't fail the request - journey updates are non-blocking
			m.l.Warn("failed to update CLI installed journey step",
				zap.String("account_id", acct.ID),
				zap.String("user_agent", userAgent),
				zap.Error(err))
		}
	}
}

// isCLIUserAgent checks if the User-Agent string indicates CLI usage
func isCLIUserAgent(userAgent string) bool {
	userAgent = strings.ToLower(userAgent)

	// Check for various patterns that would indicate CLI usage
	cliIndicators := []string{
		"nuon-cli",       // Explicit CLI identifier
		"nuon/",          // Version pattern like "nuon/1.2.3"
		"go-http-client", // Go's default HTTP client (commonly used by CLI tools)
		"curl",           // User using curl directly
		"wget",           // User using wget
		"postman",        // Postman client (some users use this for testing CLI endpoints)
	}

	for _, indicator := range cliIndicators {
		if strings.Contains(userAgent, indicator) {
			return true
		}
	}

	return false
}
