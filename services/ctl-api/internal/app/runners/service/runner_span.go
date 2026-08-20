package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

var runnerTracer = otel.Tracer("github.com/nuonco/nuon/services/ctl-api/internal/app/runners/service")

func traceRunnerRequestOperation(ctx *gin.Context, name string, fn func(context.Context) error) error {
	request := ctx.Request
	spanCtx, span := runnerTracer.Start(request.Context(), name)
	ctx.Request = request.WithContext(spanCtx)
	defer func() {
		ctx.Request = request
		span.End()
	}()

	operationCtx := trace.ContextWithSpan(ctx, span)
	if err := fn(operationCtx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

func setRunnerSpanAttributes(ctx *gin.Context, runner *app.Runner, installID, installName string, extra ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx.Request.Context())
	if !span.IsRecording() {
		return
	}

	attrs := make([]attribute.KeyValue, 0, 6+len(extra))
	attrs = append(attrs,
		attribute.String("nuon.org_id", runner.OrgID),
		attribute.String("nuon.org_name", runner.Org.Name),
		attribute.String("nuon.runner_id", runner.ID),
		attribute.String("nuon.runner_type", string(runner.RunnerGroup.Type)),
		attribute.String("nuon.install_id", installID),
		attribute.String("nuon.install_name", installName),
	)
	attrs = append(attrs, extra...)
	span.SetAttributes(attrs...)
}
