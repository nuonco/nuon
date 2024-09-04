package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): we have to define this here, because the `ptraceotlp.ExportRequest` type is actually a hidden type and
// means we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://opentelemetry.io/docs/specs/otel/protocol/file-exporter/#examples
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
			Scope struct{} `json:"scope"`
			Spans []struct {
				TraceID                string `json:"traceId"`
				SpanID                 string `json:"spanId"`
				ParentSpanID           string `json:"parentSpanId"`
				Name                   string `json:"name"`
				StartTimeUnixNano      string `json:"startTimeUnixNano"`
				EndTimeUnixNano        string `json:"endTimeUnixNano"`
				DroppedAttributesCount int    `json:"droppedAttributesCount,omitempty"`
				Events                 []struct {
					TimeUnixNano string `json:"timeUnixNano"`
					Name         string `json:"name"`
					Attributes   []struct {
						Key   string `json:"key"`
						Value struct {
							StringValue string `json:"stringValue"`
						} `json:"value"`
					} `json:"attributes,omitempty"`
					DroppedAttributesCount int `json:"droppedAttributesCount"`
				} `json:"events,omitempty"`
				DroppedEventsCount int `json:"droppedEventsCount,omitempty"`
				Status             struct {
					Message string `json:"message"`
					Code    int    `json:"code"`
				} `json:"status,omitempty"`
				Links []struct {
					TraceID    string `json:"traceId"`
					SpanID     string `json:"spanId"`
					Attributes []struct {
						Key   string `json:"key"`
						Value struct {
							StringValue string `json:"stringValue"`
						} `json:"value"`
					} `json:"attributes,omitempty"`
					DroppedAttributesCount int `json:"droppedAttributesCount"`
				} `json:"links,omitempty"`
				DroppedLinksCount int      `json:"droppedLinksCount,omitempty"`
				Status0           struct{} `json:"status,omitempty"`
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

	var req *ptraceotlp.ExportRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	results, err := s.writeRunnerTraces(ctx, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to write runner traces: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, results)
}

func (s *service) writeRunnerTraces(ctx context.Context, runnerID string, req *ptraceotlp.ExportRequest) ([]*app.RunnerJob, error) {
	return nil, nil
}
