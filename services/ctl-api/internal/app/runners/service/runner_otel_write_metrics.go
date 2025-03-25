package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/otel"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
)

//	@ID						RunnerOtelWriteMetrics
//	@Summary				runner write metrics
//	@Description.markdown	runner_otel_write_metrics.md
//	@Param					runner_id	path	string		true	"runner ID"
//	@Param					req			body	interface{}	true	"Input"
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
//	@Router					/v1/runners/{runner_id}/metrics [POST]
func (s *service) OtelWriteMetrics(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	// read data into bytes
	jsonData, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	// unmarshal bytes into ExportRequest
	// NOTE(fd): this is essentially our validation step. we do not use this object directly otherwise.
	expreq := pmetricotlp.NewExportRequest()
	if err := expreq.UnmarshalJSON(jsonData); err != nil {
		ctx.Error(fmt.Errorf("unable to marshal request: %w", err))
		return
	}

	writerrs := s.writeRunnerMetrics(ctx, runnerID, &expreq)
	if writerrs != nil {
		for _, err := range writerrs {
			ctx.Error(fmt.Errorf("unable to write runner metrics: %w", err))
		}
		return
	}

	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) writeRunnerMetrics(ctx context.Context, runnerID string, req *pmetricotlp.ExportRequest) []error {

	// a list of errors to be returned
	errors := []error{}

	// list of metric objects to be created
	sumMetrics := []app.OtelMetricSumIngestion{}
	gaugeMetrics := []app.OtelMetricGaugeIngestion{}
	histogramMetrics := []app.OtelMetricHistogramIngestion{}
	exponentialHistogramMetrics := []app.OtelMetricExponentialHistogramIngestion{}
	summaryMetrics := []app.OtelMetricSummaryIngestion{}

	metricsSlice := req.Metrics().ResourceMetrics()
	for i := 0; i < metricsSlice.Len(); i++ {
		metrics := metricsSlice.At(i)

		resAttr := metrics.Resource().Attributes()
		resourceAttrsMap := otel.AttributesToMap(resAttr)

		resourceSchemaUrl := metrics.SchemaUrl()

		// NOTE(fd): this is a well established convention.
		var serviceName string
		snVal, ok := resAttr.Get("service.name")
		if ok {
			serviceName = snVal.AsString()
		}

		// NOTE(fd): this is a nuon convention.
		var runnerJobExecutionId string
		runnerJobExecutionVal, ok := resAttr.Get("runner_job_execution.id")
		if ok {
			runnerJobExecutionId = runnerJobExecutionVal.AsString()
			fmt.Printf("runnerJobExecutionId=%s", runnerJobExecutionId)
		}

		scopeMetrics := metrics.ScopeMetrics()
		for j := 0; j < scopeMetrics.Len(); j++ {
			rs := scopeMetrics.At(j).Metrics()
			scopeInstr := scopeMetrics.At(j).Scope()
			scopeAttrMap := otel.AttributesToMap(scopeInstr.Attributes())
			scopeURL := scopeMetrics.At(j).SchemaUrl()

			for k := 0; k < rs.Len(); k++ {
				r := rs.At(k)

				switch r.Type() {
				case pmetric.MetricTypeGauge:
					for l := 0; l < r.Gauge().DataPoints().Len(); l++ {
						dp := r.Gauge().DataPoints().At(i)
						dpAttrs := otel.AttributesToMap(dp.Attributes())
						attrs, times, values, traceIDs, spanIDs := otel.ConvertExemplars(dp.Exemplars())
						gaugeMetrics = append(gaugeMetrics, app.OtelMetricGaugeIngestion{
							RunnerID:               runnerID,
							RunnerGroupID:          resourceAttrsMap["runner_group.id"],
							RunnerJobID:            dpAttrs["runner_job.id"],
							RunnerJobExecutionID:   dpAttrs["runner_job_execution.id"],
							RunnerJobExecutionStep: dpAttrs["runner_job_execution_step.name"],

							ResourceAttributes: resourceAttrsMap,
							ResourceSchemaURL:  resourceSchemaUrl,

							ScopeName:             scopeInstr.Name(),
							ScopeVersion:          scopeInstr.Version(),
							ScopeAttributes:       scopeAttrMap,
							ScopeDroppedAttrCount: int(scopeInstr.DroppedAttributesCount()),
							ScopeSchemaURL:        scopeURL,

							ServiceName: serviceName,

							MetricName:        r.Name(),
							MetricDescription: r.Description(),
							MetricUnit:        r.Unit(),

							StartTimeUnix: dp.StartTimestamp().AsTime(),
							TimeUnix:      dp.Timestamp().AsTime(),
							Value:         otel.GetValue(dp.IntValue(), dp.DoubleValue(), dp.ValueType()),
							Flags:         uint32(dp.Flags()),

							ExemplarsFilteredAttributes: attrs,
							ExemplarsTimeUnix:           times,
							ExemplarsValue:              values,
							ExemplarsTraceID:            traceIDs,
							ExemplarsSpanID:             spanIDs,
						})
					}
				case pmetric.MetricTypeSum:
					for l := 0; l < r.Sum().DataPoints().Len(); l++ {
						dp := r.Sum().DataPoints().At(i)
						dpAttrs := otel.AttributesToMap(dp.Attributes())
						attrs, times, values, traceIDs, spanIDs := otel.ConvertExemplars(dp.Exemplars())
						sumMetrics = append(sumMetrics, app.OtelMetricSumIngestion{
							RunnerID:               runnerID,
							RunnerGroupID:          resourceAttrsMap["runner_group.id"],
							RunnerJobID:            dpAttrs["runner_job.id"],
							RunnerJobExecutionID:   dpAttrs["runner_job_execution.id"],
							RunnerJobExecutionStep: dpAttrs["runner_job_execution_step.name"],

							ResourceAttributes: otel.AttributesToMap(resAttr),
							ResourceSchemaURL:  resourceSchemaUrl,

							ScopeName:             scopeInstr.Name(),
							ScopeVersion:          scopeInstr.Version(),
							ScopeAttributes:       otel.AttributesToMap(scopeInstr.Attributes()),
							ScopeDroppedAttrCount: int(scopeInstr.DroppedAttributesCount()),
							ScopeSchemaURL:        scopeURL,

							ServiceName: serviceName,

							MetricName:        r.Name(),
							MetricDescription: r.Description(),
							MetricUnit:        r.Unit(),

							StartTimeUnix: dp.StartTimestamp().AsTime(),
							TimeUnix:      dp.Timestamp().AsTime(),
							Value:         otel.GetValue(dp.IntValue(), dp.DoubleValue(), dp.ValueType()),
							Flags:         uint32(dp.Flags()),

							IsMonotonic: r.Sum().IsMonotonic(),

							AggregationTemporality:      int32(r.Sum().AggregationTemporality()),
							ExemplarsFilteredAttributes: attrs,
							ExemplarsTimeUnix:           times,
							ExemplarsValue:              values,
							ExemplarsTraceID:            traceIDs,
							ExemplarsSpanID:             spanIDs,
						})
					}
				case pmetric.MetricTypeHistogram:
					for l := 0; l < r.Histogram().DataPoints().Len(); l++ {
						dp := r.Histogram().DataPoints().At(i)
						dpAttrs := otel.AttributesToMap(dp.Attributes())
						attrs, times, values, traceIDs, spanIDs := otel.ConvertExemplars(dp.Exemplars())
						histogramMetrics = append(histogramMetrics, app.OtelMetricHistogramIngestion{
							RunnerID:               runnerID,
							RunnerGroupID:          resourceAttrsMap["runner_group.id"],
							RunnerJobID:            dpAttrs["runner_job.id"],
							RunnerJobExecutionID:   dpAttrs["runner_job_execution.id"],
							RunnerJobExecutionStep: dpAttrs["runner_job_execution_step.name"],

							ResourceAttributes: otel.AttributesToMap(resAttr),
							ResourceSchemaURL:  resourceSchemaUrl,

							ScopeName:             scopeInstr.Name(),
							ScopeVersion:          scopeInstr.Version(),
							ScopeAttributes:       otel.AttributesToMap(scopeInstr.Attributes()),
							ScopeDroppedAttrCount: uint32(scopeInstr.DroppedAttributesCount()),
							ScopeSchemaURL:        scopeURL,

							ServiceName: serviceName,

							MetricName:        r.Name(),
							MetricDescription: r.Description(),
							MetricUnit:        r.Unit(),

							StartTimeUnix: dp.StartTimestamp().AsTime(),
							TimeUnix:      dp.Timestamp().AsTime(),

							Count:          dp.Count(),
							Sum:            dp.Sum(),
							BucketsCount:   otel.ConvertSliceToArraySet(dp.BucketCounts().AsRaw()),
							ExplicitBounds: otel.ConvertSliceToArraySet(dp.ExplicitBounds().AsRaw()),

							Flags: uint32(dp.Flags()),
							Min:   dp.Min(),
							Max:   dp.Max(),

							AggregationTemporality: int32(r.Histogram().AggregationTemporality()),

							ExemplarsFilteredAttributes: attrs,
							ExemplarsTimeUnix:           times,
							ExemplarsValue:              values,
							ExemplarsTraceID:            traceIDs,
							ExemplarsSpanID:             spanIDs,
						})
					}
				case pmetric.MetricTypeExponentialHistogram:
					for l := 0; l < r.ExponentialHistogram().DataPoints().Len(); l++ {
						dp := r.ExponentialHistogram().DataPoints().At(l)
						dpAttrs := otel.AttributesToMap(dp.Attributes())
						attrs, times, values, traceIDs, spanIDs := otel.ConvertExemplars(dp.Exemplars())
						exponentialHistogramMetrics = append(exponentialHistogramMetrics, app.OtelMetricExponentialHistogramIngestion{
							RunnerID:               runnerID,
							RunnerGroupID:          resourceAttrsMap["runner_group.id"],
							RunnerJobID:            dpAttrs["runner_job.id"],
							RunnerJobExecutionID:   dpAttrs["runner_job_execution.id"],
							RunnerJobExecutionStep: dpAttrs["runner_job_execution_step.name"],

							ResourceAttributes: otel.AttributesToMap(resAttr),
							ResourceSchemaURL:  resourceSchemaUrl,

							ScopeName:             scopeInstr.Name(),
							ScopeVersion:          scopeInstr.Version(),
							ScopeAttributes:       otel.AttributesToMap(scopeInstr.Attributes()),
							ScopeDroppedAttrCount: uint32(scopeInstr.DroppedAttributesCount()),
							ScopeSchemaURL:        scopeURL,

							ServiceName: serviceName,

							MetricName:        r.Name(),
							MetricDescription: r.Description(),
							MetricUnit:        r.Unit(),

							StartTimeUnix: dp.StartTimestamp().AsTime(),
							TimeUnix:      dp.Timestamp().AsTime(),

							Count: dp.Count(),
							Sum:   dp.Sum(),

							Scale:                dp.Scale(),
							ZeroCount:            dp.ZeroCount(),
							PositiveOffset:       dp.Positive().Offset(),
							PositiveBucketCounts: otel.ConvertSliceToArraySet(dp.Positive().BucketCounts().AsRaw()),
							NegativeOffset:       dp.Negative().Offset(),
							NegativeBucketCounts: otel.ConvertSliceToArraySet(dp.Negative().BucketCounts().AsRaw()),

							Flags: uint32(dp.Flags()),
							Min:   dp.Min(),
							Max:   dp.Max(),

							AggregationTemporality: int32(r.ExponentialHistogram().AggregationTemporality()),

							ExemplarsFilteredAttributes: attrs,
							ExemplarsTimeUnix:           times,
							ExemplarsValue:              values,
							ExemplarsTraceID:            traceIDs,
							ExemplarsSpanID:             spanIDs,
						})
					}
				case pmetric.MetricTypeSummary:
					for l := 0; l < r.Summary().DataPoints().Len(); l++ {
						dp := r.Summary().DataPoints().At(l)
						dpAttrs := otel.AttributesToMap(dp.Attributes())
						quantiles, values := otel.ConvertValueAtQuantile(dp.QuantileValues())
						summaryMetrics = append(summaryMetrics, app.OtelMetricSummaryIngestion{
							RunnerID:               runnerID,
							RunnerGroupID:          resourceAttrsMap["runner_group.id"],
							RunnerJobID:            dpAttrs["runner_job.id"],
							RunnerJobExecutionID:   dpAttrs["runner_job_execution.id"],
							RunnerJobExecutionStep: dpAttrs["runner_job_execution_step.name"],

							ResourceAttributes: otel.AttributesToMap(resAttr),
							ResourceSchemaURL:  resourceSchemaUrl,

							ScopeName:             scopeInstr.Name(),
							ScopeVersion:          scopeInstr.Version(),
							ScopeAttributes:       otel.AttributesToMap(scopeInstr.Attributes()),
							ScopeDroppedAttrCount: int(scopeInstr.DroppedAttributesCount()),
							ScopeSchemaURL:        scopeURL,

							ServiceName: serviceName,

							MetricName:        r.Name(),
							MetricDescription: r.Description(),
							MetricUnit:        r.Unit(),

							StartTimeUnix: dp.StartTimestamp().AsTime(),
							TimeUnix:      dp.Timestamp().AsTime(),

							Count:                    dp.Count(),
							Sum:                      dp.Sum(),
							Flags:                    uint32(dp.Flags()),
							ValueAtQuantilesQuantile: quantiles,
							ValueAtQuantilesValue:    values,
						})
					}
				case pmetric.MetricTypeEmpty:
					// TODO: log and consider returning early (or ... validate these up top)
					errors = append(errors, fmt.Errorf("metrics type is unset"))

				default:
					// TODO: log and consider returning early
					errors = append(errors, fmt.Errorf("unsupported metrics type"))
				}
			}
		}
	}

	// TODO: collect all errors in a list and return the list of errors
	if len(sumMetrics) > 0 {
		sumRes := s.chDB.WithContext(ctx).Create(&sumMetrics)
		if sumRes.Error != nil {
			errors = append(errors, fmt.Errorf("unable to ingest sum metrics: %w", sumRes.Error))
		}
	}

	if len(gaugeMetrics) > 0 {
		gaugeRes := s.chDB.WithContext(ctx).Create(&gaugeMetrics)
		if gaugeRes.Error != nil {
			errors = append(errors, fmt.Errorf("unable to ingest gauge metrics: %w", gaugeRes.Error))
		}
	}

	if len(histogramMetrics) > 0 {
		histogramRes := s.chDB.WithContext(ctx).Create(&histogramMetrics)
		if histogramRes.Error != nil {
			errors = append(errors, fmt.Errorf("unable to ingest histogram metrics: %w", histogramRes.Error))
		}
	}

	if len(exponentialHistogramMetrics) > 0 {
		exponentialHistogramRes := s.chDB.WithContext(ctx).Create(&exponentialHistogramMetrics)
		if exponentialHistogramRes.Error != nil {
			errors = append(errors, fmt.Errorf("unable to ingest exponential histogram metrics: %w", exponentialHistogramRes.Error))
		}
	}

	if len(summaryMetrics) > 0 {
		summaryRes := s.chDB.WithContext(ctx).Create(&summaryMetrics)
		if summaryRes.Error != nil {
			errors = append(errors, fmt.Errorf("unable to ingest summary metrics: %w", summaryRes.Error))
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}
