package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/workflow"
)

const (
	orgIDCtxKey string = "org_id"
)

func OrgIDFromContext(ctx ValueContext) (string, error) {
	orgID := ctx.Value(orgIDCtxKey)
	if orgID == nil {
		return "", fmt.Errorf("org was not set on middleware context")
	}

	return orgID.(string), nil
}

func SetOrgIDGinContext(ctx *gin.Context, orgID string) {
	ctx.Set(orgIDCtxKey, orgID)
}

func SetOrgIDContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, orgIDCtxKey, orgID)
}

func SetOrgIDWorkflowContext(ctx workflow.Context, orgID string) workflow.Context {
	return workflow.WithValue(ctx, orgIDCtxKey, orgID)
}