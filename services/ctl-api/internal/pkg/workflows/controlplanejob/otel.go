package controlplanejob

import (
	"context"
	"fmt"
	"time"

	runnercontrolplane "github.com/nuonco/nuon/pkg/runner/controlplane"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/kafka"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (a *Activities) WriteControlPlaneLogs(ctx context.Context, logStreamID string, records []runnercontrolplane.OTELLogRecord) error {
	if len(records) == 0 {
		return nil
	}

	logStream, err := a.getLogStream(ctx, logStreamID)
	if err != nil {
		return err
	}
	ctx = cctx.SetOrgIDContext(ctx, logStream.OrgID)
	ctx = cctx.SetAccountIDContext(ctx, logStream.CreatedByID)

	logStreamIDs := []string{logStream.ID}
	if !logStream.ParentLogStreamID.Empty() {
		logStreamIDs = append(logStreamIDs, logStream.ParentLogStreamID.String)
	}

	// One receive time for the batch, so the parent fan-out matches the child.
	now := time.Now()

	logs := make([]app.OtelLogRecord, 0, len(records)*len(logStreamIDs))
	for _, targetLogStreamID := range logStreamIDs {
		for _, record := range records {
			logs = append(logs, app.OtelLogRecord{
				// Stamped here rather than left to BeforeCreate / GORM autofill,
				// which resolve at insert time: on the Kafka path that happens in a
				// consumer with no request context, so org_id would be empty (it
				// leads the destination table's ORDER BY) and created_at would mean
				// "when the sink flushed" instead of "when we got this".
				ID:          domains.NewOtelLogID(),
				OrgID:       logStream.OrgID,
				CreatedByID: logStream.CreatedByID,
				CreatedAt:   now,
				UpdatedAt:   now,

				LogStreamID:            targetLogStreamID,
				RunnerID:               record.RunnerID,
				RunnerGroupID:          record.RunnerGroupID,
				RunnerJobID:            record.RunnerJobID,
				RunnerJobExecutionID:   record.RunnerJobExecutionID,
				RunnerJobExecutionStep: record.RunnerJobExecutionStep,
				ResourceAttributes:     record.ResourceAttributes,
				ResourceSchemaURL:      record.ResourceSchemaURL,
				ScopeSchemaURL:         record.ScopeSchemaURL,
				ScopeName:              record.ScopeName,
				ScopeVersion:           record.ScopeVersion,
				ScopeAttributes:        record.ScopeAttributes,
				Timestamp:              record.Timestamp,
				TimestampTime:          record.Timestamp,
				TimestampDate:          record.Timestamp,
				ServiceName:            record.ServiceName,
				SeverityNumber:         record.SeverityNumber,
				SeverityText:           record.SeverityText,
				Body:                   record.Body,
				TraceID:                record.TraceID,
				SpanID:                 record.SpanID,
				TraceFlags:             record.TraceFlags,
				LogAttributes:          record.LogAttributes,
			})
		}
	}

	if !a.kafka.Enabled() {
		return a.writeControlPlaneLogs(ctx, logs)
	}

	// Synchronous for the same reason as the runner OTLP path: this activity's
	// current write is a blocking ClickHouse insert, and Temporal treats a
	// returning activity as durably done. Producing fire-and-forget would let the
	// activity succeed while the records only exist in a process buffer, and a
	// worker restart would lose them with nothing left to retry.
	msgs := make([]kafka.Message, 0, len(logs))
	for _, log := range logs {
		msgs = append(msgs, kafka.Message{Key: log.LogStreamID, Payload: log})
	}

	failed := a.kafka.ProduceEnvelopesSync(ctx, kafka.TopicOtelLogRecords, kafka.TypeOtelLogRecord, msgs)
	if len(failed) == 0 {
		return nil
	}

	fallback := make([]app.OtelLogRecord, 0, len(failed))
	for _, i := range failed {
		fallback = append(fallback, logs[i])
	}

	a.l.Warn("unable to produce control-plane logs to kafka; writing unacked records inline",
		zap.Int("acked", len(logs)-len(failed)),
		zap.Int("fallback", len(fallback)),
	)

	return a.writeControlPlaneLogs(ctx, fallback)
}

func (a *Activities) writeControlPlaneLogs(ctx context.Context, logs []app.OtelLogRecord) error {
	if err := a.chDB.WithContext(ctx).Create(&logs).Error; err != nil {
		return fmt.Errorf("unable to ingest control-plane logs: %w", err)
	}
	return nil
}

func (a *Activities) WriteControlPlaneTraces(ctx context.Context, runnerID string, records []runnercontrolplane.OTELTraceRecord) error {
	if len(records) == 0 {
		return nil
	}

	ctx = a.traceWriteContext(ctx, records)

	// Stamped here rather than left to BeforeCreate / GORM autofill, which
	// resolve at insert time: on the Kafka path that happens in a consumer with
	// no request context, so created_by_id would be empty and created_at would
	// mean "when the sink flushed" instead of "when we got this".
	createdByID := keys.CreatedByIDFromContext(ctx)
	now := time.Now()

	traces := make([]app.OtelTraceIngestion, 0, len(records))
	for _, record := range records {
		recordRunnerID := record.RunnerID
		if recordRunnerID == "" {
			recordRunnerID = runnerID
		}
		traces = append(traces, app.OtelTraceIngestion{
			ID:          domains.NewOtelTraceID(),
			CreatedByID: createdByID,
			CreatedAt:   now,
			UpdatedAt:   now,

			RunnerID:               recordRunnerID,
			RunnerGroupID:          record.RunnerGroupID,
			RunnerJobID:            record.RunnerJobID,
			RunnerJobExecutionID:   record.RunnerJobExecutionID,
			RunnerJobExecutionStep: record.RunnerJobExecutionStep,
			Timestamp:              record.Timestamp,
			TimestampTime:          record.Timestamp,
			TimestampDate:          record.Timestamp,
			ResourceAttributes:     record.ResourceAttributes,
			ResourceSchemaURL:      record.ResourceSchemaURL,
			ScopeSchemaURL:         record.ScopeSchemaURL,
			ScopeName:              record.ScopeName,
			ScopeVersion:           record.ScopeVersion,
			ScopeAttributes:        record.ScopeAttributes,
			TraceID:                record.TraceID,
			SpanID:                 record.SpanID,
			ParentSpanID:           record.ParentSpanID,
			TraceState:             record.TraceState,
			SpanName:               record.SpanName,
			SpanKind:               record.SpanKind,
			ServiceName:            record.ServiceName,
			SpanAttributes:         record.SpanAttributes,
			Duration:               record.Duration,
			StatusCode:             record.StatusCode,
			StatusMessage:          record.StatusMessage,
			EventsTimestamp:        record.EventsTimestamp,
			EventsName:             record.EventsName,
			EventsAttributes:       record.EventsAttributes,
			LinksTraceID:           record.LinksTraceID,
			LinksSpanID:            record.LinksSpanID,
			LinksState:             record.LinksState,
			LinksAttributes:        record.LinksAttributes,
		})
	}

	if !a.kafka.Enabled() {
		return a.writeControlPlaneTraces(ctx, traces)
	}

	// Synchronous for the same reason as the control-plane logs path above:
	// Temporal treats a returning activity as durably done, so a fire-and-forget
	// produce would let this activity succeed while the spans only exist in a
	// process buffer, and a worker restart would lose them with nothing left to
	// retry.
	msgs := make([]kafka.Message, 0, len(traces))
	for _, trace := range traces {
		key := trace.RunnerJobID
		if key == "" {
			key = trace.RunnerID
		}
		msgs = append(msgs, kafka.Message{Key: key, Payload: trace})
	}

	failed := a.kafka.ProduceEnvelopesSync(ctx, kafka.TopicOtelTraces, kafka.TypeOtelTrace, msgs)
	if len(failed) == 0 {
		return nil
	}

	fallback := make([]app.OtelTraceIngestion, 0, len(failed))
	for _, i := range failed {
		fallback = append(fallback, traces[i])
	}

	a.l.Warn("unable to produce control-plane traces to kafka; writing unacked spans inline",
		zap.Int("acked", len(traces)-len(failed)),
		zap.Int("fallback", len(fallback)),
	)

	return a.writeControlPlaneTraces(ctx, fallback)
}

func (a *Activities) writeControlPlaneTraces(ctx context.Context, traces []app.OtelTraceIngestion) error {
	if err := a.chDB.WithContext(ctx).Create(&traces).Error; err != nil {
		return fmt.Errorf("unable to ingest control-plane traces: %w", err)
	}
	return nil
}

func (a *Activities) traceWriteContext(ctx context.Context, records []runnercontrolplane.OTELTraceRecord) context.Context {
	for _, record := range records {
		if record.RunnerJobID == "" {
			continue
		}
		job, err := a.getJob(ctx, record.RunnerJobID)
		if err != nil {
			return ctx
		}
		ctx = cctx.SetOrgIDContext(ctx, job.OrgID)
		ctx = cctx.SetAccountIDContext(ctx, job.CreatedByID)
		return ctx
	}
	return ctx
}

func (a *Activities) getLogStream(ctx context.Context, logStreamID string) (*app.LogStream, error) {
	var logStream app.LogStream
	err := a.db.WithContext(ctx).
		Where(&app.LogStream{ID: logStreamID}).
		First(&logStream).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("control-plane log stream %s not found", logStreamID)
		}
		return nil, fmt.Errorf("unable to get control-plane log stream: %w", err)
	}
	return &logStream, nil
}
