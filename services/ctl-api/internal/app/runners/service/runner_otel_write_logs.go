package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/collector/pdata/plog/plogotlp"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): we have to define this here, because the `plogotlp.ExportRequest` type is actually a hidden type and means
// we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://opentelemetry.io/docs/specs/otel/protocol/file-exporter/#examples
type OTLPLogExportRequest struct {
	ResourceLogs []struct {
		Resource struct {
			Attributes []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue string `json:"stringValue"`
				} `json:"value"`
			} `json:"attributes"`
		} `json:"resource"`
		ScopeLogs []struct {
			Scope      struct{} `json:"scope"`
			LogRecords []struct {
				TimeUnixNano   string `json:"timeUnixNano"`
				SeverityNumber int    `json:"severityNumber"`
				SeverityText   string `json:"severityText"`
				Body           struct {
					StringValue string `json:"stringValue"`
				} `json:"body"`
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value,omitempty"`
					Value0 struct {
						IntValue string `json:"intValue"`
					} `json:"value,omitempty"`
				} `json:"attributes"`
				DroppedAttributesCount int    `json:"droppedAttributesCount"`
				TraceID                string `json:"traceId"`
				SpanID                 string `json:"spanId"`
			} `json:"logRecords"`
		} `json:"scopeLogs"`
	} `json:"resourceLogs"`
}

// @ID RunnerOtelWriteLogs
// @Summary	runner write logs
// @Description.markdown runner_otel_write_logs.md
// @Param			runner_id	path	string	true	"runner ID"
// @Param			req				body	OTLPLogExportRequest true	"Input"
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
// @Router			/v1/runners/{runner_id}/logs [POST]
func (s *service) OtelWriteLogs(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req *plogotlp.ExportRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err := s.writeRunnerLogs(ctx, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to write runner logs: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) writeRunnerLogs(ctx context.Context, runnerID string, logs *plogotlp.ExportRequest) ([]*app.RunnerJob, error) {
	return nil, nil
}
