package auth

import (
	"fmt"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/agentclient"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	runnershelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/useragent"
)

type Params struct {
	fx.In

	L               *zap.Logger
	MW              metrics.Writer
	Cfg             *internal.Config
	DB              *gorm.DB `name:"psql"`
	AuthzClient     *authz.Client
	AcctClient      *account.Client
	AccountsHelpers *accountshelpers.Helpers
	RunnersHelpers  *runnershelpers.Helpers
}

type middleware struct {
	cfg             *internal.Config
	l               *zap.Logger
	mw              metrics.Writer
	db              *gorm.DB
	authzClient     *authz.Client
	acctClient      *account.Client
	accountsHelpers *accountshelpers.Helpers
	runnersHelpers  *runnershelpers.Helpers
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

		// fetchAccountToken returns (nil, nil) when the token is not found in the DB.
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
			cctx.SetTokenTypeGinContext(ctx, acctToken.TokenType)
			m.detectCLIUsage(ctx, acct)
			m.recordPersonalStaticToken(ctx, acct, acctToken)
			m.logAgentRequest(ctx, acct.ID)
			ctx.Next()
			return
		}

		ctx.Error(stderr.ErrAuthentication{
			Err:         fmt.Errorf("token not recognized"),
			Description: "Please make sure you are using a valid token.",
		})
		ctx.Abort()
	}
}

// detectCLIUsage checks if the request is from CLI and updates the journey step
func (m *middleware) detectCLIUsage(ctx *gin.Context, acct *app.Account) {
	userAgent := ctx.Request.UserAgent()
	if !useragent.IsCLI(userAgent) {
		return
	}

	if err := m.accountsHelpers.UpdateUserJourneyStepForCLIInstalled(ctx, acct.ID); err != nil {
		m.l.Warn("failed to update cli_installed journey step",
			zap.String("account_id", acct.ID),
			zap.Error(err),
		)
	}
}

func (m *middleware) recordPersonalStaticToken(ctx *gin.Context, acct *app.Account, token *app.Token) {
	if token.TokenType != app.TokenTypeStatic || acct.AccountType == app.AccountTypeService {
		return
	}

	m.l.Info("personal static token used",
		zap.String("account_id", acct.ID),
		zap.String("token_id", token.ID),
		zap.String("org_id", ctx.GetHeader("X-Nuon-Org-ID")),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	)
	m.mw.Incr("auth.personal_static_token", []string{
		"method:" + ctx.Request.Method,
		"path:" + ctx.FullPath(),
	})
}

func (m *middleware) logAgentRequest(ctx *gin.Context, accountID string) {
	agent := ctx.GetHeader(agentclient.Header)
	if agent == "" {
		return
	}
	m.l.Info("nuon cli agent",
		zap.String("agent", agent),
		zap.String("command", ctx.GetHeader(agentclient.CommandHeader)),
		zap.String("account_id", accountID),
		zap.String("method", ctx.Request.Method),
		zap.String("path", ctx.FullPath()),
	)
}

func (m *middleware) Name() string {
	return "auth"
}

func New(params Params) *middleware {
	return &middleware{
		l:               params.L,
		mw:              params.MW,
		cfg:             params.Cfg,
		db:              params.DB,
		authzClient:     params.AuthzClient,
		acctClient:      params.AcctClient,
		accountsHelpers: params.AccountsHelpers,
		runnersHelpers:  params.RunnersHelpers,
	}
}
