package service

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/collector/pdata/plog/plogotlp"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/utils"
)

// NOTE(jm): we have to define this here, because the `plogotlp.ExportRequest` type is actually a hidden type and means
// we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://github.com/open-telemetry/opentelemetry-proto/blob/main/examples/logs.json#L67

// NOTE(fd): Attributes can be key, and StringValue, IntValue, BoolValue, ArrayValue, KeylistValue, etc.
// this struct is not used for validation of incoming or outgoing data, it is just used in our API docs.
// validation takes place when we unmarsal the json w/ expreq.UnmarshalJSON.

type Attribute struct {
	Key   string `json:"key"`
	Value struct {
		StringValue string `json:"stringValue"`
	} `json:"value,omitempty"`
	Value0 struct {
		BoolValue bool `json:"boolValue"`
	} `json:"value,omitempty"`
	Value1 struct {
		IntValue string `json:"intValue"`
	} `json:"value,omitempty"`
	Value2 struct {
		DoubleValue float64 `json:"doubleValue"`
	} `json:"value,omitempty"`
	Value3 struct {
		ArrayValue struct {
			Values []struct {
				StringValue string `json:"stringValue"`
			} `json:"values"`
		} `json:"arrayValue"`
	} `json:"value,omitempty"`
	Value4 struct {
		KvlistValue struct {
			Values []struct {
				Key   string `json:"key"`
				Value struct {
					StringValue string `json:"stringValue"`
				} `json:"value"`
			} `json:"values"`
		} `json:"kvlistValue"`
	} `json:"value,omitempty"`
}
type Resource struct {
	Attributes []Attribute `json:"attributes"`
}

type Scope struct {
	Name                   string      `json:"name,omitempty"`
	Version                string      `json:"version,omitempty"`
	Attributes             []Attribute `json:"attributes,omitempty"`
	DroppedAttributesCount uint32      `json:"droppedAttributesCount,omitempty"`
}

type Body struct {
	StringValue string `json:"stringValue"`
}

type OTLPLogExportRequest struct {
	ResourceLogs []struct {
		Resource  `json:"resource"`
		ScopeLogs []struct {
			SchemaURL  string `json:"schemaUrl,omitempty"`
			Scope      Scope  `json:"scope"`
			LogRecords []struct {
				TimeUnixNano           string      `json:"timeUnixNano"`
				SeverityNumber         int         `json:"severityNumber"`
				SeverityText           string      `json:"severityText"`
				ServiceName            string      `json:"serviceName"`
				Flags                  int         `json:"flags,omitempty"`
				DroppedAttributesCount int         `json:"droppedAttributesCount"`
				TraceID                string      `json:"traceId"`
				SpanID                 string      `json:"spanId"`
				Body                   Body        `json:"body"`
				Attributes             []Attribute `json:"attributes"`
			} `json:"logRecords"`
		} `json:"scopeLogs"`
	} `json:"resourceLogs"`
}

// NOTE(jm): we have to define this here, because the `plogotlp.ExportRequest` type is actually a hidden type and means
// we would have to define this otherwise.
//
// Instead, we use https://mholt.github.io/json-to-go/ to generate the types from the example JSON in the OTEL examples
// here: https://opentelemetry.io/docs/specs/otel/protocol/file-exporter/#examples

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

	// write the logs to the db
	writeErr := s.writeRunnerLogs(ctx, runnerID, expreq)
	if writeErr != nil {
		ctx.Error(fmt.Errorf("unable to write runner logs: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) writeRunnerLogs(ctx context.Context, runnerID string, logs plogotlp.ExportRequest) error {
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
		resourceSchemaUrl := log.SchemaUrl()

		// NOTE(fd): this is a well established convention.
		var serviceName string
		snVal, ok := resourceAttributes.Get("service.name")
		if ok {
			serviceName = snVal.AsString()
		}

		// NOTE(fd): this is a nuon convention.
		var jobId string
		jobIdVal, ok := resourceAttributes.Get("runner_job.id")
		if ok {
			jobId = jobIdVal.AsString()
		}

		// NOTE(fd): this is a nuon convention.
		var runnerGroupId string
		runnerGroupIdVal, ok := resourceAttributes.Get("runner_group.id")
		if ok {
			runnerGroupId = runnerGroupIdVal.AsString()
		}

		// NOTE(fd): this is a nuon convention.
		var runnerJobExecutionId string
		runnerJobExecutionVal, ok := resourceAttributes.Get("runner_job_execution.id")
		if ok {
			runnerJobExecutionId = runnerJobExecutionVal.AsString()
		}

		scopeLogs := log.ScopeLogs()

		for j := 0; j < scopeLogs.Len(); j++ {
			scopeLog := scopeLogs.At(j)
			scopeAttrs := scopeLog.Scope().Attributes()
			scopeName := scopeLog.Scope().Name()
			scopeVersion := scopeLog.Scope().Version()
			scopeSchemaUrl := scopeLog.SchemaUrl()
			logRecords := scopeLog.LogRecords()
			for k := 0; k < logRecords.Len(); k++ {
				log := logRecords.At(k)
				timestamp := log.Timestamp().AsTime()
				logAttrs := log.Attributes()

				otelLogRecords = append(otelLogRecords, app.OtelLogRecord{
					// runner info
					RunnerID:             runnerID,
					RunnerJobID:          jobId,
					RunnerGroupID:        runnerGroupId,
					RunnerJobExecutionID: runnerJobExecutionId,

					// from resource
					ResourceAttributes: utils.AttributesToMap(resourceAttrs),
					ResourceSchemaURL:  resourceSchemaUrl,

					// from scope
					ScopeSchemaURL:  scopeSchemaUrl,
					ScopeName:       scopeName,
					ScopeVersion:    scopeVersion,
					ScopeAttributes: utils.AttributesToMap(scopeAttrs),

					Timestamp:      timestamp,
					TimestampTime:  timestamp, // the gorm model struct sets these to zero so we must be explici
					TimestampDate:  timestamp, // the gorm model struct sets these to zero so we must be explici
					ServiceName:    serviceName,
					SeverityNumber: int(log.SeverityNumber()),
					SeverityText:   log.SeverityText(),
					Body:           log.Body().AsString(),
					TraceID:        log.TraceID().String(),
					SpanID:         log.SpanID().String(),
					TraceFlags:     int(log.Flags()),
					LogAttributes:  utils.AttributesToMap(logAttrs),
				})
			}
		}
	}

	// write the otel logs to the db
	res := s.chDB.WithContext(ctx).Create(&otelLogRecords)
	if res.Error != nil {
		return fmt.Errorf("unable to ingest logs: %w", res.Error)
	}
	// save to db
	return nil
}
