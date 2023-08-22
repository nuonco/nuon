package auth

import (
	"fmt"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

type middleware struct {
	cfg *internal.Config
	l   *zap.Logger
	db  *gorm.DB
}

func (m *middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := jwtmiddleware.AuthHeaderTokenExtractor(ctx.Request)
		if err != nil {
			ctx.Error(stderr.ErrUser{
				Err:         err,
				Description: "Please make sure you set the -H Auth:Bearer <token> header",
			})
			ctx.Abort()
			return
		}

		userToken, err := m.fetchUserToken(ctx, token)
		if err != nil {
			ctx.Error(err)
			ctx.Abort()
			return
		}
		if userToken != nil {
			ctx.Set(userTokenCtxKey, userToken)
			ctx.Set(userIDCtxKey, userToken.Subject)
			ctx.Next()
			return
		}

		// new token, so validate the token
		claims, err := m.validateToken(ctx, token)
		if err != nil {
			ctx.Error(stderr.ErrUser{
				Err:         err,
				Description: "Please sure the token is valid",
			})
			ctx.Abort()
			return
		}

		// store the token
		userToken, err = m.saveUserToken(ctx, token, claims)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to save user token: %w", err))
			ctx.Abort()
			return
		}

		ctx.Set(userTokenCtxKey, userToken)
		ctx.Set(userIDCtxKey, userToken.Subject)
		ctx.Next()
	}
}

func (m *middleware) Name() string {
	return "auth"
}

func New(l *zap.Logger, cfg *internal.Config, db *gorm.DB) *middleware {
	return &middleware{
		l:   l,
		cfg: cfg,
		db:  db,
	}
}
