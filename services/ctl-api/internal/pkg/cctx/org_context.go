package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	orgCtxKey string = "org"
)

func OrgFromContext(ctx ValueContext) (*app.Org, error) {
	org := ctx.Value(orgCtxKey)
	if org == nil {
		return nil, fmt.Errorf("org was not set on middleware context")
	}

	return org.(*app.Org), nil
}

func SetOrgGinContext(ctx *gin.Context, org *app.Org) {
	ctx.Set(orgCtxKey, org)
	ctx.Set(orgIDCtxKey, org.ID)
}

func SetOrgContext(ctx context.Context, org *app.Org) context.Context {
	ctx = context.WithValue(ctx, orgIDCtxKey, org.ID)
	return context.WithValue(ctx, orgCtxKey, org)
}
