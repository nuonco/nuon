package cctx

import (
	"context"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

type WorkflowTelemetry = keys.WorkflowTelemetry

func WorkflowTelemetryFromContext(ctx ValueContext) WorkflowTelemetry {
	telemetry, _ := ctx.Value(keys.WorkflowTelemetryCtxKey).(WorkflowTelemetry)
	return telemetry
}

func SetWorkflowTelemetryContext(ctx context.Context, telemetry WorkflowTelemetry) context.Context {
	merged := WorkflowTelemetryFromContext(ctx).Merge(telemetry)
	return context.WithValue(ctx, keys.WorkflowTelemetryCtxKey, merged)
}

func SetWorkflowTelemetryWorkflowContext(ctx workflow.Context, telemetry WorkflowTelemetry) workflow.Context {
	merged := WorkflowTelemetryFromContext(ctx).Merge(telemetry)
	return workflow.WithValue(ctx, keys.WorkflowTelemetryCtxKey, merged)
}
