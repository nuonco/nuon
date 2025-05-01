package cctx

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

func OrgIDFromContext(ctx ValueContext) (string, error) {
	orgID := ctx.Value(keys.OrgIDCtxKey)
	if orgID == nil {
		return "", fmt.Errorf("org was not set on middleware context")
	}

	return orgID.(string), nil
}

func SetOrgIDGinContext(ctx *gin.Context, orgID string) {
	ctx.Set(keys.OrgIDCtxKey, orgID)
}

func SetOrgIDContext(ctx context.Context, orgID string) context.Context {
	return context.WithValue(ctx, keys.OrgIDCtxKey, orgID)
}

func SetOrgIDWorkflowContext(ctx workflow.Context, orgID string) workflow.Context {
	return workflow.WithValue(ctx, keys.OrgIDCtxKey, orgID)
}
