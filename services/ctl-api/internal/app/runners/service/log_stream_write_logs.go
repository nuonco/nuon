package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/otel"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
)

// @ID						LogStreamWriteLogs
// @Summary				log stream write logs
// @Description.markdown	log_stream_write_logs.md
// @Param					log_stream_id	path	string						true	"log stream ID"
// @Param					req				body	otel.OTLPLogExportRequest	true	"Input"
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
// @Success				201	{object}	app.EmptyResponse
// @Router					/v1/log-streams/{log_stream_id}/logs [POST]
func (s *service) LogStreamWriteLogs(ctx *gin.Context) {
	logStreamID := ctx.Param("log_stream_id")

	byts, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	expreq := plogotlp.NewExportRequest()
	if err := expreq.UnmarshalProto(byts); err != nil {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("unable to unmarshal request: %w", err)))
		return
	}

	logStream, err := s.getCachedLogStream(ctx, logStreamID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get log stream"))
		return
	}

	s.mw.Incr("otel.log_stream.batch", metrics.ToTags(map[string]string{
		"log_stream_type": logStream.OwnerType,
	}))
	s.mw.Gauge("otel.log_stream.batch_size", float64(len(byts)), metrics.ToTags(map[string]string{
		"log_stream_type": logStream.OwnerType,
	}))

	now := time.Now()

	logs := s.toLogStreamLogs(ctx, now, logStreamID, expreq)

	if !logStream.ParentLogStreamID.Empty() {
		logs = append(logs, s.toLogStreamLogs(ctx, now, logStream.ParentLogStreamID.String, expreq)...)
	}

	err = s.produceOrWriteLogStreamLogs(ctx, logs)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to write runner logs: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, app.EmptyResponse{})
}

func (s *service) toLogStreamLogs(ctx context.Context, now time.Time, logStreamID string, logs plogotlp.ExportRequest) []app.OtelLogRecord {
	orgID := keys.OrgIDFromContext(ctx)
	createdByID := keys.CreatedByIDFromContext(ctx)

	otelLogRecords := []app.OtelLogRecord{}

	logSlice := logs.Logs().ResourceLogs()
	for i := 0; i < logSlice.Len(); i++ {
		log := logSlice.At(i)

		resourceAttributes := log.Resource().Attributes()
		resourceAttrs := resourceAttributes
		resourceAttrsMap := otel.AttributesToMap(resourceAttrs)
		resourceSchemaUrl := log.SchemaUrl()

		var resourceServiceName string
		snVal, ok := resourceAttributes.Get("service.name")
		if ok {
			resourceServiceName = snVal.AsString()
		}

		scopeLogs := log.ScopeLogs()

		for j := 0; j < scopeLogs.Len(); j++ {
			scopeLog := scopeLogs.At(j)
			scopeAttrs := scopeLog.Scope().Attributes()
			scopeAttrMap := otel.AttributesToMap(scopeAttrs)
			scopeName := scopeLog.Scope().Name()
			scopeVersion := scopeLog.Scope().Version()
			scopeSchemaUrl := scopeLog.SchemaUrl()
			logRecords := scopeLog.LogRecords()
			for k := 0; k < logRecords.Len(); k++ {
				log := logRecords.At(k)
				timestamp := log.Timestamp().AsTime()
				logAttrs := log.Attributes()
				logAttributesMap := otel.AttributesToMap(logAttrs)

				serviceName := resourceServiceName
				if v, ok := logAttributesMap["service.name"]; ok && v != "" {
					serviceName = v
				}

				otelLogRecords = append(otelLogRecords, app.OtelLogRecord{
					ID:          domains.NewOtelLogID(),
					OrgID:       orgID,
					CreatedByID: createdByID,
					// why: Receive time, not sink-insert time. GORM would otherwise
					// autofill these when the row is written, which on the Kafka
					// path happens in the consumer — turning "when we got this log
					// line" into "when the sink flushed it", offset by the fetch
					// interval.
					CreatedAt: now,
					UpdatedAt: now,

					LogStreamID:            logStreamID,
					RunnerID:               generics.FindMap("runner.id", logAttributesMap, resourceAttrsMap),
					RunnerGroupID:          resourceAttrsMap["runner_group.id"],
					RunnerJobID:            generics.FindMap("runner_job.id", logAttributesMap, resourceAttrsMap),
					RunnerJobExecutionID:   generics.FindMap("runner_job_execution.id", logAttributesMap, resourceAttrsMap),
					RunnerJobExecutionStep: generics.FindMap("runner_job_execution_step.name", logAttributesMap, resourceAttrsMap),

					ResourceAttributes: otel.AttributesToMap(resourceAttrs),
					ResourceSchemaURL:  resourceSchemaUrl,

					ScopeSchemaURL:  scopeSchemaUrl,
					ScopeName:       scopeName,
					ScopeVersion:    scopeVersion,
					ScopeAttributes: scopeAttrMap,

					Timestamp:      timestamp,
					TimestampTime:  timestamp, // why: the gorm model struct sets these to zero so we must be explici
					TimestampDate:  timestamp, // why: the gorm model struct sets these to zero so we must be explici
					ServiceName:    serviceName,
					SeverityNumber: int(log.SeverityNumber()),
					SeverityText:   log.SeverityNumber().String(),
					Body:           log.Body().AsString(),
					TraceID:        log.TraceID().String(),
					SpanID:         log.SpanID().String(),
					TraceFlags:     int(log.Flags()),
					LogAttributes:  logAttributesMap,
				})
			}
		}
	}

	return otelLogRecords
}

// why: produceOrWriteLogStreamLogs hands the records to Kafka when it's enabled,
// falling back to the inline ClickHouse write for anything Kafka didn't ack.
//
// Synchronous, unlike the heartbeat producer. This handler currently blocks on a
// ClickHouse insert before returning 201, so the runner's success response means
// "durably stored". Producing fire-and-forget would quietly weaken that to
// "buffered in this process", and an OOM kill would drop log lines that nothing
// upstream knows to resend. Heartbeats can be async because their existing path
// is already a lossy in-memory buffer; logs have no such slack.
//
// One envelope per record rather than one per export request: the topic's
// max.message.bytes is 4MiB and a single OTLP export can carry hundreds of
// records, so per-record keeps us clear of a broker reject. They still go in a
// single ProduceSync call, so this is one round trip, not one per record.
func (s *service) produceOrWriteLogStreamLogs(ctx context.Context, logs []app.OtelLogRecord) error {
	if len(logs) == 0 {
		return nil
	}

	if !s.kafka.Enabled() {
		return s.writeLogStreamLogs(ctx, logs)
	}

	msgs := make([]kafka.Message, 0, len(logs))
	for _, log := range logs {
		msgs = append(msgs, kafka.Message{Key: log.LogStreamID, Payload: log})
	}

	failed := s.kafka.ProduceEnvelopesSync(ctx, kafka.TopicOtelLogRecords, kafka.TypeOtelLogRecord, msgs)
	if len(failed) == 0 {
		return nil
	}

	fallback := make([]app.OtelLogRecord, 0, len(failed))
	for _, i := range failed {
		fallback = append(fallback, logs[i])
	}

	s.l.Warn("unable to produce otel logs to kafka; writing unacked records inline",
		zap.Int("acked", len(logs)-len(failed)),
		zap.Int("fallback", len(fallback)),
	)

	return s.writeLogStreamLogs(ctx, fallback)
}

func (s *service) writeLogStreamLogs(ctx context.Context, logs []app.OtelLogRecord) error {
	res := s.chDB.WithContext(ctx).
		Create(&logs)
	if res.Error != nil {
		return fmt.Errorf("unable to ingest logs: %w", res.Error)
	}

	return nil
}
