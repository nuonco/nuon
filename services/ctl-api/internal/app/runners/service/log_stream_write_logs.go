package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	chClause "github.com/powertoolsdev/mono/pkg/gorm/clickhouse/pkg/clause"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/otel"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
)

// @ID LogStreamWriteLogs
// @Summary	log stream write logs
// @Description.markdown log_stream_write_logs.md
// @Param			log_stream_id	path	string	true	"log stream ID"
// @Param			req				body	otel.OTLPLogExportRequest true	"Input"
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
// @Router			/v1/log-streams/{log_stream_id}/logs [POST]
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
		ctx.Error(fmt.Errorf("unable to unmarshal request: %w", err))
		return
	}

	logStream, err := s.getLogStream(ctx, logStreamID)
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

	// write the logs to the db
	logs := s.toLogStreamLogs(logStreamID, expreq)

	if !logStream.ParentLogStreamID.Empty() {
		logs = append(logs, s.toLogStreamLogs(logStream.ParentLogStreamID.String, expreq)...)
	}

	err = s.writeLogStreamLogs(ctx, logs)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to write runner logs: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) toLogStreamLogs(logStreamID string, logs plogotlp.ExportRequest) []app.OtelLogRecord {
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
		var serviceName string
		snVal, ok := resourceAttributes.Get("service.name")
		if ok {
			serviceName = snVal.AsString()
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

				otelLogRecords = append(otelLogRecords, app.OtelLogRecord{
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

func (s *service) writeLogStreamLogs(ctx context.Context, logs []app.OtelLogRecord) error {
	// write the otel logs to the db
	res := s.chDB.WithContext(ctx).
		Clauses(chClause.AsyncInsert{}).
		Create(&logs)
	if res.Error != nil {
		return fmt.Errorf("unable to ingest logs: %w", res.Error)
	}

	// save to db
	return nil
}
