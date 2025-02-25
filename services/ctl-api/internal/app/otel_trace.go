package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OtelTraceEvent struct {
	Timestamp  time.Time         `json:"timestamp"`
	Name       string            `json:"name"`
	Attributes map[string]string `json:"attributes"`
}

type OtelTraceLink struct {
	TraceID    string            `json:"trace_id"`
	SpanID     string            `json:"span_id"`
	SpanState  string            `json:"span_state"`
	Attributes map[string]string `json:"attributes"`
}

type OtelTrace struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `gorm:"notnull;" json:"created_by_id"`

	CreatedAt time.Time             `gorm:"notnull" json:"created_at" `
	UpdatedAt time.Time             `gorm:"notnull" json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	RunnerID               string `json:"runner_id" `
	RunnerJobID            string `json:"runner_job_id"`
	RunnerGroupID          string `json:"runner_group_id"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step"`

	// OTEL log trace attributes
	Timestamp     time.Time `json:"timestamp" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1);"`
	TimestampDate time.Time `json:"timestamp_date" gorm:"type:Date;default:toDate(timestamp);"`
	TimestampTime time.Time `json:"timestamp_time" gorm:"type:DateTime;default:toDateTime(timestamp);"`

	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1);"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ScopeDroppedAttrCount int               `json:"scope_dropped_attr_count"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`

	TraceID        string            `json:"trace_id" gorm:"codec:ZSTD(1);index:idx_trace_id,type:bloom_filter(0.001),granularity:1;"`
	SpanID         string            `json:"span_id" gorm:"codec:ZSTD(1);"`
	ParentSpanID   string            `json:"parent_span_id" gorm:"codec:ZSTD(1);"`
	TraceState     string            `json:"trace_state" gorm:"codec:ZSTD(1);"`
	SpanName       string            `json:"span_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	SpanKind       string            `json:"span_kind" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	ServiceName    string            `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	SpanAttributes map[string]string `json:"span_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_span_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_span_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	Duration       int64             `json:"duration" gorm:"codec:ZSTD(1);"`
	StatusCode     string            `json:"status_code" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	StatusMessage  string            `json:"status_message" gorm:"codec:ZSTD(1);"`

	// Nested Fields
	// NOTE(fd): these control the actual migration. careful when modifying. ALTER does not work the same way on nested clickhouse columns.
	Events []OtelTraceEvent `gorm:"type:Nested(timestamp DateTime64(9), name LowCardinality(String), attributes Map(LowCardinality(String), String));"`
	Links  []OtelTraceLink  `gorm:"type:Nested(trace_id String, span_id String, span_state String, attributes Map(LowCardinality(String), String));"`
}

func (r OtelTrace) GetTableOptions() string {
	opts := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/otel_traces', '{replica}')
	TTL toDateTime("timestamp") + toIntervalDay(720)
	PARTITION BY toDate(timestamp)
	PRIMARY KEY (runner_id, runner_job_id, runner_group_id, runner_job_execution_id)
	ORDER BY    (runner_id, runner_job_id, runner_group_id, runner_job_execution_id, toUnixTimestamp(timestamp), span_name, trace_id)
	SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;`
	return opts
}

func (r OtelTrace) GetTableClusterOptions() string {
	return "on cluster simple"
}

func (r *OtelTrace) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewOtelTraceID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

// NOTE(fd): DO NOT MIGRATE THIS
// it's just here so we can write to the table and read data w/ Nested columns as an Array of Structs
type OtelTraceIngestion struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `gorm:"notnull;" json:"created_by_id"`

	CreatedAt time.Time             `gorm:"notnull" json:"created_at" `
	UpdatedAt time.Time             `gorm:"notnull" json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	RunnerID               string `json:"runner_id" `
	RunnerJobID            string `json:"runner_job_id"`
	RunnerGroupID          string `json:"runner_group_id"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step"`

	// OTEL log trace attributes
	Timestamp     time.Time `json:"timestamp" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1);"`
	TimestampDate time.Time `json:"timestamp_date" gorm:"type:Date;default:toDate(timestamp);"`
	TimestampTime time.Time `json:"timestamp_time" gorm:"type:DateTime;default:toDateTime(timestamp);"`

	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1);"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ScopeDroppedAttrCount int               `json:"scope_dropped_attr_count"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`

	TraceID        string            `json:"trace_id" gorm:"codec:ZSTD(1);index:idx_trace_id,type:bloom_filter(0.001),granularity:1;"`
	SpanID         string            `json:"span_id" gorm:"codec:ZSTD(1);"`
	ParentSpanID   string            `json:"parent_span_id" gorm:"codec:ZSTD(1);"`
	TraceState     string            `json:"trace_state" gorm:"codec:ZSTD(1);"`
	SpanName       string            `json:"span_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	SpanKind       string            `json:"span_kind" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	ServiceName    string            `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	SpanAttributes map[string]string `json:"span_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_span_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_span_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	Duration       int64             `json:"duration" gorm:"codec:ZSTD(1);"`
	StatusCode     string            `json:"status_code" gorm:"type:LowCardinality(String);codec:ZSTD(1);"`
	StatusMessage  string            `json:"status_message" gorm:"codec:ZSTD(1);"`

	// NOTE(fd): it may be useful to scan the nested columns back into Structs in AfterFind

	// the items of interest here are these attrs/columns that define a `column` in the `gorm` struct tag so gorm knows what column to send these to
	EventsTimestamp  []time.Time         `json:"-" gorm:"type:DateTime64(9);column:events.timestamp"`
	EventsName       []string            `json:"-" gorm:"type:LowCardinality(String);column:events.name"`
	EventsAttributes []map[string]string `json:"-" gorm:"type:Map(LowCardinality(String), String);column:events.attributes"`
	LinksTraceID     []string            `json:"-" gorm:"type:LowCardinality(String);column:links.trace_id"`
	LinksSpanID      []string            `json:"-" gorm:"type:LowCardinality(String);column:links.span_id"`
	LinksState       []string            `json:"-" gorm:"type:LowCardinality(String);column:links.span_state"`
	LinksAttributes  []map[string]string `json:"-" gorm:"type:Map(LowCardinality(String), String);column:links.attributes"`
}

// TableName
func (m OtelTraceIngestion) TableName() string {
	return "otel_traces"
}

func (r *OtelTraceIngestion) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewOtelTraceID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
