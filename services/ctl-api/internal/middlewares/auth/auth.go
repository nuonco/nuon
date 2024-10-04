package auth

import (
	"fmt"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
)

type Params struct {
	fx.In

	L           *zap.Logger
	Cfg         *internal.Config
	DB          *gorm.DB `name:"psql"`
	AuthzClient *authz.Client
	AcctClient  *account.Client
}

type middleware struct {
	cfg         *internal.Config
	l           *zap.Logger
	db          *gorm.DB
	authzClient *authz.Client
	acctClient  *account.Client
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if middlewares.IsPublic(ctx) {
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

		if token == "" {
			ctx.Error(stderr.ErrAuthentication{
				Err:         fmt.Errorf("auth token was empty"),
				Description: "Please make sure you set the -H Authorization:Bearer <token> header",
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
			acct, err := m.authzClient.FetchAccount(ctx, acctToken.AccountID)
			if err != nil {
				ctx.Error(err)
				ctx.Abort()
				return
			}

			middlewares.SetGinContext(ctx, acct)
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

		acct, err := m.authzClient.FetchAccount(ctx, acctToken.AccountID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to fetch: %w", err))
			ctx.Abort()
			return
		}

		middlewares.SetGinContext(ctx, acct)
		ctx.Next()
	}
}

func (m *middleware) Name() string {
	return "auth"
}

func New(params Params) *middleware {
	return &middleware{
		l:           params.L,
		cfg:         params.Cfg,
		db:          params.DB,
		authzClient: params.AuthzClient,
		acctClient:  params.AcctClient,
	}
}
