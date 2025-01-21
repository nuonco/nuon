package middlewares

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	orgCtxKey   string = "org"
	orgIDCtxKey string = "org_id"
)

func OrgIDFromContext(ctx context.Context) (string, error) {
	orgID := ctx.Value(orgIDCtxKey)
	if orgID == nil {
		return "", fmt.Errorf("org id was not set on middleware context")
	}

	return orgID.(string), nil
}

func OrgFromContext(ctx context.Context) (*app.Org, error) {
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

func SetOrgIDGinContext(ctx *gin.Context, orgID string) {
	ctx.Set(orgIDCtxKey, orgID)
}

func SetOrgContext(ctx context.Context, org *app.Org) context.Context {
	ctx = context.WithValue(ctx, orgIDCtxKey, org.ID)
	return context.WithValue(ctx, orgCtxKey, org)
}

func SetOrgIDContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, orgIDCtxKey, orgID)
}
