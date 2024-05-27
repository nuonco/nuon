package org

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	orgIDHeaderKey string = "X-Nuon-Org-ID"
	orgCtxKey      string = "org"
	orgIDCtxKey    string = "org_id"
)

func FromContext(ctx context.Context) (*app.Org, error) {
	org := ctx.Value(orgCtxKey)
	if org == nil {
		return nil, fmt.Errorf("org was not set on middleware context")
	}

	return org.(*app.Org), nil
}

func SetGinContext(ctx *gin.Context, org *app.Org) {
	ctx.Set(orgCtxKey, org)
	ctx.Set(orgIDCtxKey, org.ID)
}

func SetContext(ctx context.Context, org *app.Org) context.Context {
	ctx = context.WithValue(ctx, orgIDCtxKey, org.ID)
	return context.WithValue(ctx, orgCtxKey, org)
}
