package cctx

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

func TokenTypeFromContext(ctx ValueContext) app.TokenType {
	tokenType, _ := ctx.Value(keys.TokenTypeCtxKey).(app.TokenType)
	return tokenType
}

func IsFederatedToken(ctx ValueContext) bool {
	return TokenTypeFromContext(ctx) == app.TokenTypeFederated
}

func SetTokenTypeGinContext(ctx *gin.Context, tokenType app.TokenType) {
	ctx.Set(keys.TokenTypeCtxKey, tokenType)
}

func SetTokenTypeContext(ctx context.Context, tokenType app.TokenType) context.Context {
	return context.WithValue(ctx, keys.TokenTypeCtxKey, tokenType)
}
