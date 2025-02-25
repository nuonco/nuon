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
	ID          string `json:"id" gorm:"primary_key"`
	CreatedByID string `json:"created_by_id" gorm:"notnull"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	OrgID                  string `json:"org_id"`
	RunnerID               string `json:"runner_id"`
	LogStreamID            string `json:"log_stream_id"`
	RunnerJobID            string `json:"runner_job_id"`
	RunnerGroupID          string `json:"runner_group_id"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step"`

	// OTEL log message attributes
	Timestamp          time.Time         `json:"timestamp" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1)"`
	TimestampDate      time.Time         `json:"timestamp_date" gorm:"type:Date;default:toDate(timestamp)"`
	TimestampTime      time.Time         `json:"timestamp_time" gorm:"type:DateTime;default:toDateTime(timestamp)"`
	TraceID            string            `json:"trace_id" gorm:"codec:ZSTD(1);index:idx_trace_id,type:bloom_filter(0.001),granularity:1;"`
	SpanID             string            `json:"span_id" gorm:"codec:ZSTD(1)"`
	TraceFlags         int               `json:"trace_flags" gorm:"type:UInt8"`
	SeverityText       string            `json:"severity_text" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	SeverityNumber     int               `json:"severity_number" gorm:"type:UInt8"`
	ServiceName        string            `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	Body               string            `json:"body" gorm:"codecZSTD(1);index:idx_body,type:tokenbf_v1(32768\\,3\\,0),granularity:8;"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ScopeSchemaURL     string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ScopeName          string            `json:"scope_name" gorm:"codec:ZSTD(1)"`
	ScopeVersion       string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ScopeAttributes    map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	LogAttributes      map[string]string `json:"log_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1); index:idx_log_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_log_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
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
