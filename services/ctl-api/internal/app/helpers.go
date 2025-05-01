package app

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx/keys"
)

const (
	HeaderInstallWorkflowID = "X-Nuon-Install-Workflow-ID"
)

// createdByIDFromTemporalContext
func createdByIDFromTemporalContext(ctx workflow.Context) string {
	val := ctx.Value(keys.AccountIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}

// createdByIDFromContext returns the user id from the context. Notably, this depends on the `middlewares/auth` to set
// this, but we do not use that code to prevent a cycle import
func createdByIDFromContext(ctx context.Context) string {
	val := ctx.Value(keys.AccountIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}

// orgIDFromContext returns the org id from the context. Notably, this depends on the `middlewares/org` to set
// this, but we do not use that code to prevent a cycle import
func orgIDFromContext(ctx context.Context) string {
	val := ctx.Value(keys.OrgIDCtxKey)
	valStr, ok := val.(string)
	if !ok {
		return ""
	}

	return valStr
}

// logStreamIDFromContext returns the org id from the context. Notably, this depends on the `middlewares/org` to set
// this, but we do not use that code to prevent a cycle import
func logstreamIDFromContext(ctx context.Context) string {
	val := ctx.Value(keys.LogStreamCtxKey)
	valObj, ok := val.(*LogStream)
	if !ok {
		return ""
	}

	return valObj.ID
}

func installWorkflowFromContext(ctx context.Context) *InstallWorkflowContext {
	val := ctx.Value(keys.WorkflowCtxKey)
	valObj, ok := val.(*InstallWorkflowContext)
	if !ok {
		return nil
	}

	return valObj
}

func configFromContext(ctx context.Context) *internal.Config {
	val := ctx.Value(keys.CfgCtxKey)
	valObj, ok := val.(*internal.Config)
	if !ok {
		return nil
	}

	return valObj
}
