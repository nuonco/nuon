package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func OrgFromContext(ctx ValueContext) (*app.Org, error) {
	org := ctx.Value(keys.OrgCtxKey)
	if org == nil {
		return nil, fmt.Errorf("org was not set on middleware context")
	}

	return org.(*app.Org), nil
}

func SetOrgGinContext(ctx *gin.Context, org *app.Org) {
	ctx.Set(keys.OrgCtxKey, org)
	ctx.Set(keys.OrgIDCtxKey, org.ID)
}

func SetOrgContext(ctx context.Context, org *app.Org) context.Context {
	ctx = context.WithValue(ctx, keys.OrgIDCtxKey, org.ID)
	return context.WithValue(ctx, keys.OrgCtxKey, org)
}
