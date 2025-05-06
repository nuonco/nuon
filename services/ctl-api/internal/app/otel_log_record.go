package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

// Logs are designed to be written via an OTLP exporter.
//
// https://opentelemetry.io/docs/specs/otel/logs/bridge-api/
//
// The clickhouse exporter, is a good reference point for this
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_logs.go
type OtelLogRecord struct {
	ID          string `json:"id,omitzero" gorm:"primary_key" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// internal attributes
	OrgID                  string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	RunnerID               string `json:"runner_id,omitzero" temporaljson:"runner_id,omitzero,omitempty"`
	LogStreamID            string `json:"log_stream_id,omitzero" temporaljson:"log_stream_id,omitzero,omitempty"`
	RunnerJobID            string `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerGroupID          string `json:"runner_group_id,omitzero" temporaljson:"runner_group_id,omitzero,omitempty"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id,omitzero" temporaljson:"runner_job_execution_id,omitzero,omitempty"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step,omitzero" temporaljson:"runner_job_execution_step,omitzero,omitempty"`

	// OTEL log message attributes
	Timestamp          time.Time         `json:"timestamp,omitzero" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1)" temporaljson:"timestamp,omitzero,omitempty"`
	TimestampDate      time.Time         `json:"timestamp_date,omitzero" gorm:"type:Date;default:toDate(timestamp)" temporaljson:"timestamp_date,omitzero,omitempty"`
	TimestampTime      time.Time         `json:"timestamp_time,omitzero" gorm:"type:DateTime;default:toDateTime(timestamp)" temporaljson:"timestamp_time,omitzero,omitempty"`
	TraceID            string            `json:"trace_id,omitzero" gorm:"codec:ZSTD(1);index:idx_trace_id,type:bloom_filter(0.001),granularity:1;" temporaljson:"trace_id,omitzero,omitempty"`
	SpanID             string            `json:"span_id,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"span_id,omitzero,omitempty"`
	TraceFlags         int               `json:"trace_flags,omitzero" gorm:"type:UInt8" temporaljson:"trace_flags,omitzero,omitempty"`
	SeverityText       string            `json:"severity_text,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"severity_text,omitzero,omitempty"`
	SeverityNumber     int               `json:"severity_number,omitzero" gorm:"type:UInt8" temporaljson:"severity_number,omitzero,omitempty"`
	ServiceName        string            `json:"service_name,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"service_name,omitzero,omitempty"`
	Body               string            `json:"body,omitzero" gorm:"codecZSTD(1);index:idx_body,type:tokenbf_v1(32768\\,3\\,0),granularity:8;" temporaljson:"body,omitzero,omitempty"`
	ResourceSchemaURL  string            `json:"resource_schema_url,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"resource_schema_url,omitzero,omitempty"`
	ResourceAttributes map[string]string `json:"resource_attributes,omitzero" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"resource_attributes,omitzero,omitempty"`
	ScopeSchemaURL     string            `json:"scope_schema_url,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_schema_url,omitzero,omitempty"`
	ScopeName          string            `json:"scope_name,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"scope_name,omitzero,omitempty"`
	ScopeVersion       string            `json:"scope_version,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_version,omitzero,omitempty"`
	ScopeAttributes    map[string]string `json:"scope_attributes,omitzero" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"scope_attributes,omitzero,omitempty"`
	LogAttributes      map[string]string `json:"log_attributes,omitzero" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1); index:idx_log_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_log_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"log_attributes,omitzero,omitempty"`
}

func (r *OtelLogRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewOtelLogID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (r OtelLogRecord) GetTableOptions() string {
	return `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/otel_log_records', '{replica}')
	TTL toDateTime("timestamp") + toIntervalDay(30)
	PARTITION BY toDate(timestamp_time)
	PRIMARY KEY  (org_id, log_stream_id, runner_job_id)
	ORDER BY     (org_id, log_stream_id ,runner_job_id, timestamp_time, timestamp)
	SETTINGS index_granularity = 8192, ttl_only_drop_parts = 0;`
}

func (r OtelLogRecord) GetTableClusterOptions() string {
	return "on cluster simple"
}
