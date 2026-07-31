package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/otel"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/zap"
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
	// Resolved here rather than left to OtelTraceIngestion's BeforeCreate hook and
	// GORM's timestamp autofill, both of which run at insert time. On the Kafka
	// path that insert happens in the consumer, where there is no request context
	// to read created_by_id from and where "now" means when the sink flushed
	// rather than when we received the span. Same values, stamped at the same
	// point for both paths, so the row is identical either way.
	createdByID := keys.CreatedByIDFromContext(ctx)
	now := time.Now()

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
				// NOTE: use Sub().Nanoseconds() — endtimestamp.Unix()-timestamp.Unix()
				// truncated sub-second spans to 0.
				duration := endtimestamp.Sub(timestamp).Nanoseconds()
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
					ID:          domains.NewOtelTraceID(),
					CreatedByID: createdByID,
					CreatedAt:   now,
					UpdatedAt:   now,

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

	return s.produceOrWriteRunnerTraces(ctx, otelTraces)
}

// produceOrWriteRunnerTraces hands the spans to Kafka when it's enabled, falling
// back to the inline ClickHouse write for anything Kafka didn't ack.
//
// Synchronous, for the same reason as the OTel logs path: this handler blocks on
// a ClickHouse insert before returning 200, so the runner's success response
// means "durably stored". Producing fire-and-forget would quietly weaken that to
// "buffered in this process", and an OOM kill would drop spans that the
// exporter's retry queue has already been told were accepted.
//
// One envelope per span rather than one per export request: the topic's
// max.message.bytes is 4MiB and a single OTLP export can carry hundreds of
// spans, each with its own attribute maps plus nested events and links, so
// per-span keeps us clear of a broker reject. They still go in a single
// ProduceSync call, so this is one round trip, not one per span.
func (s *service) produceOrWriteRunnerTraces(ctx context.Context, traces []app.OtelTraceIngestion) error {
	if len(traces) == 0 {
		return nil
	}

	if !s.kafka.Enabled() {
		return s.writeRunnerTracesInline(ctx, traces)
	}

	msgs := make([]kafka.Message, 0, len(traces))
	for _, trace := range traces {
		// Keyed by runner job so one job's spans share a partition, which keeps
		// them ordered and lands them on the same consumer. It also matches the
		// destination table's ORDER BY prefix, so a consumer's batch inserts into
		// a narrow key range instead of scattering across it. Spans produced
		// outside a job (an empty runner_job_id) fall back to the runner.
		key := trace.RunnerJobID
		if key == "" {
			key = trace.RunnerID
		}
		msgs = append(msgs, kafka.Message{Key: key, Payload: trace})
	}

	failed := s.kafka.ProduceEnvelopesSync(ctx, kafka.TopicOtelTraces, kafka.TypeOtelTrace, msgs)
	if len(failed) == 0 {
		return nil
	}

	// Only the unacked spans, so a partial failure doesn't duplicate the ones
	// Kafka already has.
	fallback := make([]app.OtelTraceIngestion, 0, len(failed))
	for _, i := range failed {
		fallback = append(fallback, traces[i])
	}

	s.l.Warn("unable to produce otel traces to kafka; writing unacked spans inline",
		zap.Int("acked", len(traces)-len(failed)),
		zap.Int("fallback", len(fallback)),
	)

	return s.writeRunnerTracesInline(ctx, fallback)
}

func (s *service) writeRunnerTracesInline(ctx context.Context, traces []app.OtelTraceIngestion) error {
	if res := s.chDB.WithContext(ctx).Create(&traces); res.Error != nil {
		return fmt.Errorf("unable to ingest traces: %w", res.Error)
	}

	return nil
}
