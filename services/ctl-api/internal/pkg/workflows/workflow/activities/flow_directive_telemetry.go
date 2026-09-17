package activities

import (
	"context"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (a *Activities) logDirective(ctx context.Context, flowEvent, directiveValue string, extra ...zap.Field) {
	if directiveValue == "" {
		return
	}

	telemetry := cctx.WorkflowTelemetryFromContext(ctx)
	fields := append([]zap.Field{
		zap.String("flow_event", flowEvent),
		zap.String("directive", directiveValue),
		zap.String("workflow_type", telemetry.WorkflowType),
		zap.String("org_name", telemetry.OrgName),
		zap.String("install_name", telemetry.InstallName),
	}, extra...)

	cctx.GetLogger(ctx, a.l).Info("flow telemetry", fields...)
}
