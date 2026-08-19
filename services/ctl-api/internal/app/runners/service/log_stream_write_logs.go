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

	// read data into bytes
	byts, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	// unmarshal bytes into ExportRequest
	// NOTE(fd): this is essentially our validation step. we do not use this object directly otherwise.
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

	// One receive time for the whole request, so the parent fan-out below stamps
	// the same value as the child rather than a few microseconds later.
	now := time.Now()

	// write the logs to the db
	logs := s.toLogStreamLogs(ctx, now, logStreamID, expreq)

	// Fan out to the parent stream, if any, so a parent's log view includes its
	// children's records. The read path matches log_stream_id exactly (it's the
	// table's sort prefix), so this duplication is what makes that read cheap.
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
	// Resolved here rather than left to app.OtelLogRecord's BeforeCreate hook,
	// which reads them off the GORM statement context. When these records are
	// produced to Kafka the insert happens in the consumer, where there is no
	// request context — and org_id leads the destination table's PRIMARY KEY and
	// ORDER BY, so an empty one collapses its sort order. Stamping them at the
	// same place for both paths also keeps the Kafka and inline writes identical.
	orgID := keys.OrgIDFromContext(ctx)
	createdByID := keys.CreatedByIDFromContext(ctx)

	// prepare a slice to hold all of the record we will be writing
	otelLogRecords := []app.OtelLogRecord{}

	// iterate over the logs in the payload
	// 1. grab the resource and extract common fields (resourceLogs.resource).
	// 2. grab the scope and extract the comman fields (resourceLogs.scopeLogs.scope)
	// 3. iterate through the resourceLogs.scopeLogs.scope.logRecords and munge it w/
	//    the shared resoruce data, scope data, and data from the request (e.g.runnerid).
	// 4. save it to clickhouse
	logSlice := logs.Logs().ResourceLogs()
	for i := 0; i < logSlice.Len(); i++ {
		log := logSlice.At(i)

		resourceAttributes := log.Resource().Attributes()
		resourceAttrs := resourceAttributes
		resourceAttrsMap := otel.AttributesToMap(resourceAttrs)
		resourceSchemaUrl := log.SchemaUrl()

		// NOTE(fd): this is a well established convention.
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

				// Allow per-record override of service.name via log attributes
				// so handlers can tag logs with finer-grained service names
				// (e.g. "runner.helm") via l.With("service.name", ...) without
				// having to construct a separate LoggerProvider per tool.
				serviceName := resourceServiceName
				if v, ok := logAttributesMap["service.name"]; ok && v != "" {
					serviceName = v
				}

				otelLogRecords = append(otelLogRecords, app.OtelLogRecord{
					ID:          domains.NewOtelLogID(),
					OrgID:       orgID,
					CreatedByID: createdByID,
					// Receive time, not sink-insert time. GORM would otherwise
					// autofill these when the row is written, which on the Kafka
					// path happens in the consumer — turning "when we got this log
					// line" into "when the sink flushed it", offset by the fetch
					// interval.
					CreatedAt: now,
					UpdatedAt: now,

					// runner info
					// NOTE(fd): these locations are a convention
					LogStreamID:            logStreamID,
					RunnerID:               generics.FindMap("runner.id", logAttributesMap, resourceAttrsMap),
					RunnerGroupID:          resourceAttrsMap["runner_group.id"],
					RunnerJobID:            generics.FindMap("runner_job.id", logAttributesMap, resourceAttrsMap),
					RunnerJobExecutionID:   generics.FindMap("runner_job_execution.id", logAttributesMap, resourceAttrsMap),
					RunnerJobExecutionStep: generics.FindMap("runner_job_execution_step.name", logAttributesMap, resourceAttrsMap),

					// from resource
					ResourceAttributes: otel.AttributesToMap(resourceAttrs),
					ResourceSchemaURL:  resourceSchemaUrl,

					// from scope
					ScopeSchemaURL:  scopeSchemaUrl,
					ScopeName:       scopeName,
					ScopeVersion:    scopeVersion,
					ScopeAttributes: scopeAttrMap,

					Timestamp:      timestamp,
					TimestampTime:  timestamp, // the gorm model struct sets these to zero so we must be explici
					TimestampDate:  timestamp, // the gorm model struct sets these to zero so we must be explici
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

// produceOrWriteLogStreamLogs hands the records to Kafka when it's enabled,
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
		// Keyed by log stream so one stream's records share a partition, which
		// keeps them ordered and lands them on the same consumer.
		msgs = append(msgs, kafka.Message{Key: log.LogStreamID, Payload: log})
	}

	failed := s.kafka.ProduceEnvelopesSync(ctx, kafka.TopicOtelLogRecords, kafka.TypeOtelLogRecord, msgs)
	if len(failed) == 0 {
		return nil
	}

	// Only the unacked records, so a partial failure doesn't duplicate the ones
	// Kafka already has.
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
	// write the otel logs to the db
	res := s.chDB.WithContext(ctx).
		Create(&logs)
	if res.Error != nil {
		return fmt.Errorf("unable to ingest logs: %w", res.Error)
	}

	// save to db
	return nil
}
