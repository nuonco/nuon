package cctx

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const resolvedAppKey = "resolved_app"

type resolvedApp struct {
	ref string
	app app.App
}

func SetResolvedAppGinContext(ctx *gin.Context, ref string, a *app.App) {
	resolved := resolvedApp{ref: ref, app: *a}
	ctx.Set(resolvedAppKey, resolved)
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), resolvedAppKey, resolved))
}

func ResolvedAppFromContext(ctx ValueContext, orgID, ref string) (*app.App, bool) {
	resolved, ok := ctx.Value(resolvedAppKey).(resolvedApp)
	if !ok || resolved.ref != ref || resolved.app.OrgID != orgID {
		return nil, false
	}
	return &resolved.app, true
}
