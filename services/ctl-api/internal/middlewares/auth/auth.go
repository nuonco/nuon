package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
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

		// fall back to the X-Nuon-Auth cookie (sent by browser SPA via credentials: 'include')
		if token == "" {
			if cookieToken, cookieErr := ctx.Cookie("X-Nuon-Auth"); cookieErr == nil {
				token = cookieToken
			}
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

			cctx.SetAccountGinContext(ctx, acct)
			m.detectCLIUsage(ctx, acct)
			ctx.Next()
			return
		}
	}
}

// isCLIUserAgent checks if the User-Agent indicates CLI usage
func isCLIUserAgent(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	cliPatterns := []string{
		"nuon-cli",
		"nuon/",
		"go-http-client",
		"curl",
		"wget",
		"postman",
	}
	for _, pattern := range cliPatterns {
		if strings.Contains(ua, pattern) {
			return true
		}
	}
	return false
}

// detectCLIUsage checks if the request is from CLI and updates the journey step
func (m *middleware) detectCLIUsage(ctx *gin.Context, acct *app.Account) {
	userAgent := ctx.Request.UserAgent()
	if !isCLIUserAgent(userAgent) {
		return
	}

	if err := m.accountsHelpers.UpdateUserJourneyStepForCLIInstalled(ctx, acct.ID); err != nil {
		m.l.Warn("failed to update cli_installed journey step",
			zap.String("account_id", acct.ID),
			zap.Error(err),
		)
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

// extractAttributionFromCookie reads marketing attribution data from the nuon_attribution cookie
// set by customer-dashboard during the auth flow. Returns nil if no attribution is present.
func (m *middleware) extractAttributionFromCookie(ctx *gin.Context) map[string]interface{} {
	cookie, err := ctx.Cookie("nuon_attribution")
	if err != nil || cookie == "" {
		return nil
	}

	var attribution map[string]interface{}
	if err := json.Unmarshal([]byte(cookie), &attribution); err != nil {
		m.l.Debug("failed to parse attribution cookie", zap.Error(err))
		return nil
	}

	return attribution
}
