package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// NOTE(jm): we have to define this here, because the `pmetricotlp.ExportRequest` type is actually a hidden type and
// means we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://opentelemetry.io/docs/specs/otel/protocol/file-exporter/#examples
type OTLPMetricExportRequest struct {
	ResourceMetrics []struct {
		Resource struct {
			Attributes []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue string `json:"stringValue"`
				} `json:"value"`
			} `json:"attributes"`
		} `json:"resource"`
		ScopeMetrics []struct {
			Scope   struct{} `json:"scope"`
			Metrics []struct {
				Name string `json:"name"`
				Unit string `json:"unit"`
				Sum  struct {
					DataPoints []struct {
						Attributes []struct {
							Key   string `json:"key"`
							Value struct {
								StringValue string `json:"stringValue"`
							} `json:"value"`
						} `json:"attributes"`
						StartTimeUnixNano string `json:"startTimeUnixNano"`
						TimeUnixNano      string `json:"timeUnixNano"`
						AsInt             string `json:"asInt"`
					} `json:"dataPoints"`
					AggregationTemporality int  `json:"aggregationTemporality"`
					IsMonotonic            bool `json:"isMonotonic"`
				} `json:"sum"`
			} `json:"metrics"`
		} `json:"scopeMetrics"`
	} `json:"resourceMetrics"`
}

// @ID RunnerOtelWriteMetrics
// @Summary	runner write metrics
// @Description.markdown runner_otel_write_metrics.md
// @Param			runner_id	path	string	true	"runner ID"
// @Param			req				body	interface{} true	"Input"
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
// @Router			/v1/runners/{runner_id}/metrics [POST]
func (s *service) OtelWriteMetrics(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req *pmetricotlp.ExportRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	_, err := s.writeRunnerMetrics(ctx, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to write runner metrics: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

func (s *service) writeRunnerMetrics(ctx context.Context, runnerID string, logs *pmetricotlp.ExportRequest) ([]*app.RunnerJob, error) {
	return nil, nil
}
