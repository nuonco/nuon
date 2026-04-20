package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/otel"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

// @ID						RunnerOtelWriteTraces
// @Summary				runner write traces
// @Description.markdown	runner_otel_write_traces.md
// @Param					runner_id	path	string						true	"runner ID"
// @Param					req			body	otel.OTLPTraceExportRequest	true	"Input"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/runners/{runner_id}/traces [POST]
func (s *service) OtelWriteTraces(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	// read data into bytes
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	var req = ptraceotlp.NewExportRequest()
	if err := req.UnmarshalJSON(jsonData); err != nil {
		ctx.Error(fmt.Errorf("unable to unmarshal request: %w", err))
		return
	}

	writeErr := s.writeRunnerTraces(ctx, runnerID, req)
	if writeErr != nil {
		ctx.Error(fmt.Errorf("unable to write runner traces: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

func (s *service) writeRunnerTraces(ctx context.Context, runnerID string, req ptraceotlp.ExportRequest) error {
	var otelTraces []app.OtelTraceIngestion
	traceSlice := req.Traces().ResourceSpans()
	for i := 0; i < traceSlice.Len(); i++ {
		trace := traceSlice.At(i)

		resourceAttributes := trace.Resource().Attributes()
		resourceAttrsMap := otel.AttributesToMap(resourceAttributes)
		resourceSchemaUrl := trace.SchemaUrl()

		var serviceName string
		val, ok := resourceAttributes.Get("service.name")
		if ok {
			serviceName = val.AsString()
		}

		scopeSpans := trace.ScopeSpans()

		for j := 0; j < scopeSpans.Len(); j++ {
			scopeSpan := scopeSpans.At(j)
			scopeAttrs := scopeSpan.Scope().Attributes()
			scopeName := scopeSpan.Scope().Name()
			scopeVersion := scopeSpan.Scope().Version()
			scopeSchemaUrl := scopeSpan.SchemaUrl()
			spans := scopeSpan.Spans()
			for k := 0; k < spans.Len(); k++ {
				span := spans.At(k)
				timestamp := span.StartTimestamp().AsTime()
				endtimestamp := span.EndTimestamp().AsTime()
				duration := endtimestamp.Unix() - timestamp.Unix()
				traceAttrs := span.Attributes()
				traceAttrsMap := otel.AttributesToMap(traceAttrs)

				eventTimes, eventNames, _ := otel.ConvertEvents(span.Events())
				eventsAttrs := make([]map[string]string, span.Events().Len())

				for ei := 0; ei < span.Events().Len(); ei++ {
					event := span.Events().At(ei)
					eventsAttrs[ei] = otel.AttributesToMap(event.Attributes())
				}

				linksTraceIDs, linksSpanIDs, linksTraceStates, _ := otel.ConvertLinks(span.Links())
				linksAttrs := make([]map[string]string, span.Links().Len())
				for li := 0; li < span.Links().Len(); li++ {
					link := span.Links().At(li)
					linksAttrs[li] = otel.AttributesToMap(link.Attributes())
				}

				obj := app.OtelTraceIngestion{
					// runner info
					RunnerID:               runnerID,
					RunnerGroupID:          resourceAttrsMap["runner_group.id"],
					RunnerJobID:            traceAttrsMap["runner_job.id"],
					RunnerJobExecutionID:   traceAttrsMap["runner_job_execution.id"],
					RunnerJobExecutionStep: traceAttrsMap["runner_job_execution_step.name"],

					// topmatter
					Timestamp:     timestamp,
					TimestampTime: timestamp,
					TimestampDate: timestamp,

					// from resource
					ResourceAttributes: resourceAttrsMap,
					ResourceSchemaURL:  resourceSchemaUrl,

					// from scope
					ScopeSchemaURL:  scopeSchemaUrl,
					ScopeName:       scopeName,
					ScopeVersion:    scopeVersion,
					ScopeAttributes: otel.AttributesToMap(scopeAttrs),

					TraceID:          span.TraceID().String(),
					SpanID:           span.SpanID().String(),
					ParentSpanID:     span.ParentSpanID().String(),
					TraceState:       span.TraceState().AsRaw(),
					SpanName:         span.Name(),
					SpanKind:         span.Kind().String(),
					ServiceName:      serviceName,
					SpanAttributes:   otel.AttributesToMap(traceAttrs),
					Duration:         duration,
					StatusCode:       span.Status().Code().String(),
					StatusMessage:    span.Status().Message(),
					EventsTimestamp:  eventTimes,
					EventsName:       eventNames,
					EventsAttributes: eventsAttrs,
					LinksTraceID:     linksTraceIDs,
					LinksSpanID:      linksSpanIDs,
					LinksState:       linksTraceStates,
					LinksAttributes:  linksAttrs,
				}

				otelTraces = append(otelTraces, obj)
			}
		}
	}

	if len(otelTraces) > 0 {
		res := s.chDB.WithContext(ctx).Create(&otelTraces)
		if res.Error != nil {
			return fmt.Errorf("unable to ingest traces: %w", res.Error)
		}
	}

	return nil
}
