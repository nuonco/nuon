package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/otel"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

//	@ID						RunnerOtelWriteTraces
//	@Summary				runner write traces
//	@Description.markdown	runner_otel_write_traces.md
//	@Param					runner_id	path	string						true	"runner ID"
//	@Param					req			body	otel.OTLPTraceExportRequest	true	"Input"
//	@Tags					runners/runner
//	@Accept					json
//	@Produce				json
//	@Security				APIKey
//	@Security				OrgID
//	@Failure				400	{object}	stderr.ErrResponse
//	@Failure				401	{object}	stderr.ErrResponse
//	@Failure				403	{object}	stderr.ErrResponse
//	@Failure				404	{object}	stderr.ErrResponse
//	@Failure				500	{object}	stderr.ErrResponse
//	@Success				201	{object}	string
//	@Router					/v1/runners/{runner_id}/traces [POST]
func (s *service) OtelWriteTraces(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	// read data into bytes
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	var req ptraceotlp.ExportRequest = ptraceotlp.NewExportRequest()
	if err := req.UnmarshalJSON(jsonData); err != nil {
		ctx.Error(fmt.Errorf("unable to unmarshal request: %w", err))
		return
	}

	writeErr := s.writeRunnerTraces(ctx, runnerID, req)
	if writeErr != nil {
		ctx.Error(fmt.Errorf("unable to write runner traces: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

func (s *service) writeRunnerTraces(ctx context.Context, runnerID string, req ptraceotlp.ExportRequest) error {

	otelTraces := []app.OtelTraceIngestion{}
	traceSlice := req.Traces().ResourceSpans()
	for i := 0; i < traceSlice.Len(); i++ {
		trace := traceSlice.At(i)

		resourceAttributes := trace.Resource().Attributes()
		resourceAttrs := resourceAttributes
		resourceAttrsMap := otel.AttributesToMap(resourceAttrs)
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
			traces := scopeSpan.Spans()
			for k := 0; k < traces.Len(); k++ {
				trace := traces.At(k)
				timestamp := trace.StartTimestamp().AsTime()
				endtimestamp := trace.EndTimestamp().AsTime()
				duration := endtimestamp.Unix() - timestamp.Unix()
				traceAttrs := trace.Attributes()
				traceAttrsMap := otel.AttributesToMap(traceAttrs)

				eventTimes, eventNames, _ := otel.ConvertEvents(trace.Events())
				eventsAttrs := make([]map[string]string, trace.Events().Len())

				for i = 0; i < trace.Events().Len(); i++ {
					event := trace.Events().At(i)
					eventsAttrs[i] = otel.AttributesToMap(event.Attributes())
				}

				linksTraceIDs, linksSpanIDs, linksTraceStates, _ := otel.ConvertLinks(trace.Links())
				linksAttrs := make([]map[string]string, trace.Links().Len())
				for i = 0; i < trace.Links().Len(); i++ {
					link := trace.Links().At(i)
					linksAttrs[i] = otel.AttributesToMap(link.Attributes())
				}

				obj := app.OtelTraceIngestion{
					// NOTE(fd): we should send the `RunnerJobID` and `RunnerJobExecutionID` in known/conventional
					// fields in the payload so we can extract them here similar to the way we extract serviceName
					// from the field: service.name. e.g. runner.job_id and runner.job_execution_id.
					// this would enable some nifty views/filtering.

					// runner info
					RunnerID:               runnerID,
					RunnerGroupID:          resourceAttrsMap["runner_group.id"],
					RunnerJobID:            traceAttrsMap["runner_job.id"],
					RunnerJobExecutionID:   traceAttrsMap["runner_job_execution.id"],
					RunnerJobExecutionStep: traceAttrsMap["runner_job_execution_step.name"],

					// topmatter
					Timestamp:     timestamp,
					TimestampTime: timestamp, // the gorm model struct sets these to zero so we must be explicit
					TimestampDate: timestamp, // the gorm model struct sets these to zero so we must be explici

					// from resource
					ResourceAttributes: resourceAttrsMap,
					ResourceSchemaURL:  resourceSchemaUrl,

					// from scope
					ScopeSchemaURL:  scopeSchemaUrl,
					ScopeName:       scopeName,
					ScopeVersion:    scopeVersion,
					ScopeAttributes: otel.AttributesToMap(scopeAttrs),

					TraceID:          trace.TraceID().String(),
					SpanID:           trace.SpanID().String(),
					ParentSpanID:     trace.ParentSpanID().String(),
					TraceState:       trace.TraceState().AsRaw(),
					SpanName:         trace.Name(),
					SpanKind:         trace.Kind().String(),
					ServiceName:      serviceName,
					SpanAttributes:   otel.AttributesToMap(traceAttrs),
					Duration:         duration,
					StatusCode:       trace.Status().Code().String(),
					StatusMessage:    trace.Status().Message(),
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

		res := s.chDB.WithContext(ctx).Create(&otelTraces)
		if res.Error != nil {
			return fmt.Errorf("unable to ingest traces: %w", res.Error)
		}

	}
	return nil
}
