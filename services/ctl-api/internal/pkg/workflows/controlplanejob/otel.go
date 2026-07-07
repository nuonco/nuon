package controlplanejob

import (
	"context"
	"fmt"

	runnercontrolplane "github.com/nuonco/nuon/pkg/runner/controlplane"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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

	logs := make([]app.OtelLogRecord, 0, len(records)*len(logStreamIDs))
	for _, targetLogStreamID := range logStreamIDs {
		for _, record := range records {
			logs = append(logs, app.OtelLogRecord{
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
	traces := make([]app.OtelTraceIngestion, 0, len(records))
	for _, record := range records {
		recordRunnerID := record.RunnerID
		if recordRunnerID == "" {
			recordRunnerID = runnerID
		}
		traces = append(traces, app.OtelTraceIngestion{
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
