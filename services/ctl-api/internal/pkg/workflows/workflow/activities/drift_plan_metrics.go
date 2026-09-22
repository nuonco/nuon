package activities

import (
	"context"
	"errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func (a *Activities) recordDriftPlanEvaluationForTarget(ctx context.Context, workflowType app.WorkflowType, installID, componentID string, isNoop bool, stage string, err error) {
	if a.driftPlanEvaluations == nil {
		return
	}
	var target string
	switch workflowType {
	case app.WorkflowTypeDriftRun:
		target = "component"
	case app.WorkflowTypeDriftRunReprovisionSandbox:
		target = "sandbox"
	default:
		return
	}
	attrs := []attribute.KeyValue{attribute.String("target", target)}
	if installID != "" {
		attrs = append(attrs, attribute.String("nuon.install.id", installID))
	}
	if target == "component" && componentID != "" {
		attrs = append(attrs, attribute.String("nuon.component.id", componentID))
	}
	if err != nil {
		outcome := "error"
		if ctx.Err() != nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			outcome = "cancelled"
		}
		attrs = append(attrs, attribute.String("outcome", outcome), attribute.String("error.type", stage))
	} else {
		decision := "drift"
		if isNoop {
			decision = "no_drift"
		}
		attrs = append(attrs, attribute.String("outcome", "success"), attribute.String("decision", decision))
	}
	a.driftPlanEvaluations.Add(ctx, 1, metric.WithAttributes(attrs...))
}
