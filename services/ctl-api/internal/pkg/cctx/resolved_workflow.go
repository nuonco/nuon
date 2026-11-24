package cctx

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const resolvedWorkflowKey = "resolved_workflow"

func SetResolvedWorkflowGinContext(ctx *gin.Context, workflow *app.Workflow) {
	ctx.Set(resolvedWorkflowKey, *workflow)
	ctx.Request = ctx.Request.WithContext(context.WithValue(ctx.Request.Context(), resolvedWorkflowKey, *workflow))
}

func ResolvedWorkflowFromContext(ctx ValueContext, orgID, workflowID string) (*app.Workflow, bool) {
	workflow, ok := ctx.Value(resolvedWorkflowKey).(app.Workflow)
	if !ok || workflow.ID != workflowID || workflow.OrgID != orgID {
		return nil, false
	}
	return &workflow, true
}
