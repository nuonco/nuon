package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/utils"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
)

// NOTE(jm): we have to define this here, because the `ptraceotlp.ExportRequest` type is actually a hidden type and
// means we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://github.com/open-telemetry/opentelemetry-proto/blob/main/examples/trace.json
type OTLPTraceExportRequest struct {
	ResourceSpans []struct {
		Resource struct {
			Attributes []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue string `json:"stringValue"`
				} `json:"value"`
			} `json:"attributes"`
		} `json:"resource"`
		ScopeSpans []struct {
			Scope struct {
				Name       string `json:"name"`
				Version    string `json:"version"`
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"scope"`
			Spans []struct {
				TraceID           string `json:"traceId"`
				SpanID            string `json:"spanId"`
				ParentSpanID      string `json:"parentSpanId"`
				Name              string `json:"name"`
				StartTimeUnixNano string `json:"startTimeUnixNano"`
				EndTimeUnixNano   string `json:"endTimeUnixNano"`
				Kind              int    `json:"kind"`
				Attributes        []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"spans"`
		} `json:"scopeSpans"`
	} `json:"resourceSpans"`
}

// @ID RunnerOtelWriteTraces
// @Summary	runner write traces
// @Description.markdown runner_otel_write_traces.md
// @Param			runner_id	path	string	true	"runner ID"
// @Param			req				body	OTLPTraceExportRequest true	"Input"
// @Tags runners/runner
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	string
// @Router			/v1/runners/{runner_id}/traces [POST]
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
				timestamp := trace.StartTimestamp().AsTime().Unix()
				endtimestamp := trace.EndTimestamp().AsTime().Unix()
				duration := endtimestamp - timestamp
				traceAttrs := trace.Attributes()

				eventTimes, eventNames, _ := utils.ConvertEvents(trace.Events())
				eventsAttrs := make([]map[string]string, trace.Events().Len())

				for i = 0; i < trace.Events().Len(); i++ {
					event := trace.Events().At(i)
					eventsAttrs[i] = utils.AttributesToMap(event.Attributes())
				}

				linksTraceIDs, linksSpanIDs, linksTraceStates, _ := utils.ConvertLinks(trace.Links())
				linksAttrs := make([]map[string]string, trace.Links().Len())
				for i = 0; i < trace.Links().Len(); i++ {
					link := trace.Links().At(i)
					linksAttrs[i] = utils.AttributesToMap(link.Attributes())
				}

				obj := app.OtelTraceIngestion{
					// NOTE(fd): we should send the `RunnerJobID` and `RunnerJobExecutionID` in known/conventional
					// fields in the payload so we can extract them here similar to the way we extract serviceName
					// from the field: service.name. e.g. runner.job_id and runner.job_execution_id.
					// this would enable some nifty views/filtering.

					// runner info
					RunnerID: runnerID,

					// topmatter
					Timestamp: time.Unix(timestamp, 0),

					// from resource
					ResourceAttributes: utils.AttributesToMap(resourceAttrs),
					ResourceSchemaURL:  resourceSchemaUrl,

					// from scope
					ScopeSchemaURL:  scopeSchemaUrl,
					ScopeName:       scopeName,
					ScopeVersion:    scopeVersion,
					ScopeAttributes: utils.AttributesToMap(scopeAttrs),

					TraceID:          trace.TraceID().String(),
					SpanID:           trace.SpanID().String(),
					ParentSpanID:     trace.ParentSpanID().String(),
					TraceState:       trace.TraceState().AsRaw(),
					SpanName:         trace.Name(),
					SpanKind:         trace.Kind().String(),
					ServiceName:      serviceName,
					SpanAttributes:   utils.AttributesToMap(traceAttrs),
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
