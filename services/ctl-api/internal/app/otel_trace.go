package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OtelTraceEvent struct {
	Timestamp  time.Time         `json:"timestamp" temporaljson:"timestamp,omitzero,omitempty"`
	Name       string            `json:"name" temporaljson:"name,omitzero,omitempty"`
	Attributes map[string]string `json:"attributes" temporaljson:"attributes,omitzero,omitempty"`
}

type OtelTraceLink struct {
	TraceID    string            `json:"trace_id" temporaljson:"trace_id,omitzero,omitempty"`
	SpanID     string            `json:"span_id" temporaljson:"span_id,omitzero,omitempty"`
	SpanState  string            `json:"span_state" temporaljson:"span_state,omitzero,omitempty"`
	Attributes map[string]string `json:"attributes" temporaljson:"attributes,omitzero,omitempty"`
}

type OtelTrace struct {
	ID          string `gorm:"primary_key" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `gorm:"notnull;" json:"created_by_id" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `gorm:"notnull" json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `gorm:"notnull" json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// internal attributes
	RunnerID               string `json:"runner_id" temporaljson:"runner_id,omitzero,omitempty"`
	RunnerJobID            string `json:"runner_job_id" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerGroupID          string `json:"runner_group_id" temporaljson:"runner_group_id,omitzero,omitempty"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id" temporaljson:"runner_job_execution_id,omitzero,omitempty"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step" temporaljson:"runner_job_execution_step,omitzero,omitempty"`

	// OTEL log trace attributes
	Timestamp     time.Time `json:"timestamp" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1);" temporaljson:"timestamp,omitzero,omitempty"`
	TimestampDate time.Time `json:"timestamp_date" gorm:"type:Date;default:toDate(timestamp);" temporaljson:"timestamp_date,omitzero,omitempty"`
	TimestampTime time.Time `json:"timestamp_time" gorm:"type:DateTime;default:toDateTime(timestamp);" temporaljson:"timestamp_time,omitzero,omitempty"`

	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"resource_attributes,omitzero,omitempty"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"resource_schema_url,omitzero,omitempty"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1);" temporaljson:"scope_name,omitzero,omitempty"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"scope_version,omitzero,omitempty"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"scope_attributes,omitzero,omitempty"`
	ScopeDroppedAttrCount int               `json:"scope_dropped_attr_count" temporaljson:"scope_dropped_attr_count,omitzero,omitempty"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"scope_schema_url,omitzero,omitempty"`

	TraceID        string            `json:"trace_id" gorm:"codec:ZSTD(1);index:idx_trace_id,type:bloom_filter(0.001),granularity:1;" temporaljson:"trace_id,omitzero,omitempty"`
	SpanID         string            `json:"span_id" gorm:"codec:ZSTD(1);" temporaljson:"span_id,omitzero,omitempty"`
	ParentSpanID   string            `json:"parent_span_id" gorm:"codec:ZSTD(1);" temporaljson:"parent_span_id,omitzero,omitempty"`
	TraceState     string            `json:"trace_state" gorm:"codec:ZSTD(1);" temporaljson:"trace_state,omitzero,omitempty"`
	SpanName       string            `json:"span_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"span_name,omitzero,omitempty"`
	SpanKind       string            `json:"span_kind" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"span_kind,omitzero,omitempty"`
	ServiceName    string            `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"service_name,omitzero,omitempty"`
	SpanAttributes map[string]string `json:"span_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_span_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_span_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"span_attributes,omitzero,omitempty"`
	Duration       int64             `json:"duration" gorm:"codec:ZSTD(1);" temporaljson:"duration,omitzero,omitempty"`
	StatusCode     string            `json:"status_code" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"status_code,omitzero,omitempty"`
	StatusMessage  string            `json:"status_message" gorm:"codec:ZSTD(1);" temporaljson:"status_message,omitzero,omitempty"`

	// Nested Fields
	// NOTE(fd): these control the actual migration. careful when modifying. ALTER does not work the same way on nested clickhouse columns.
	Events []OtelTraceEvent `gorm:"type:Nested(timestamp DateTime64(9), name LowCardinality(String), attributes Map(LowCardinality(String), String));" temporaljson:"events,omitzero,omitempty"`
	Links  []OtelTraceLink  `gorm:"type:Nested(trace_id String, span_id String, span_state String, attributes Map(LowCardinality(String), String));" temporaljson:"links,omitzero,omitempty"`
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
	ID          string `gorm:"primary_key" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `gorm:"notnull;" json:"created_by_id" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `gorm:"notnull" json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `gorm:"notnull" json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// internal attributes
	RunnerID               string `json:"runner_id" temporaljson:"runner_id,omitzero,omitempty"`
	RunnerJobID            string `json:"runner_job_id" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerGroupID          string `json:"runner_group_id" temporaljson:"runner_group_id,omitzero,omitempty"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id" temporaljson:"runner_job_execution_id,omitzero,omitempty"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step" temporaljson:"runner_job_execution_step,omitzero,omitempty"`

	// OTEL log trace attributes
	Timestamp     time.Time `json:"timestamp" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1);" temporaljson:"timestamp,omitzero,omitempty"`
	TimestampDate time.Time `json:"timestamp_date" gorm:"type:Date;default:toDate(timestamp);" temporaljson:"timestamp_date,omitzero,omitempty"`
	TimestampTime time.Time `json:"timestamp_time" gorm:"type:DateTime;default:toDateTime(timestamp);" temporaljson:"timestamp_time,omitzero,omitempty"`

	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"resource_attributes,omitzero,omitempty"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"resource_schema_url,omitzero,omitempty"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1);" temporaljson:"scope_name,omitzero,omitempty"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"scope_version,omitzero,omitempty"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"scope_attributes,omitzero,omitempty"`
	ScopeDroppedAttrCount int               `json:"scope_dropped_attr_count" temporaljson:"scope_dropped_attr_count,omitzero,omitempty"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"scope_schema_url,omitzero,omitempty"`

	TraceID        string            `json:"trace_id" gorm:"codec:ZSTD(1);index:idx_trace_id,type:bloom_filter(0.001),granularity:1;" temporaljson:"trace_id,omitzero,omitempty"`
	SpanID         string            `json:"span_id" gorm:"codec:ZSTD(1);" temporaljson:"span_id,omitzero,omitempty"`
	ParentSpanID   string            `json:"parent_span_id" gorm:"codec:ZSTD(1);" temporaljson:"parent_span_id,omitzero,omitempty"`
	TraceState     string            `json:"trace_state" gorm:"codec:ZSTD(1);" temporaljson:"trace_state,omitzero,omitempty"`
	SpanName       string            `json:"span_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"span_name,omitzero,omitempty"`
	SpanKind       string            `json:"span_kind" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"span_kind,omitzero,omitempty"`
	ServiceName    string            `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"service_name,omitzero,omitempty"`
	SpanAttributes map[string]string `json:"span_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_span_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_span_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"span_attributes,omitzero,omitempty"`
	Duration       int64             `json:"duration" gorm:"codec:ZSTD(1);" temporaljson:"duration,omitzero,omitempty"`
	StatusCode     string            `json:"status_code" gorm:"type:LowCardinality(String);codec:ZSTD(1);" temporaljson:"status_code,omitzero,omitempty"`
	StatusMessage  string            `json:"status_message" gorm:"codec:ZSTD(1);" temporaljson:"status_message,omitzero,omitempty"`

	// NOTE(fd): it may be useful to scan the nested columns back into Structs in AfterFind

	// the items of interest here are these attrs/columns that define a `column` in the `gorm` struct tag so gorm knows what column to send these to
	EventsTimestamp  []time.Time         `json:"-" gorm:"type:DateTime64(9);column:events.timestamp" temporaljson:"events_timestamp,omitzero,omitempty"`
	EventsName       []string            `json:"-" gorm:"type:LowCardinality(String);column:events.name" temporaljson:"events_name,omitzero,omitempty"`
	EventsAttributes []map[string]string `json:"-" gorm:"type:Map(LowCardinality(String), String);column:events.attributes" temporaljson:"events_attributes,omitzero,omitempty"`
	LinksTraceID     []string            `json:"-" gorm:"type:LowCardinality(String);column:links.trace_id" temporaljson:"links_trace_id,omitzero,omitempty"`
	LinksSpanID      []string            `json:"-" gorm:"type:LowCardinality(String);column:links.span_id" temporaljson:"links_span_id,omitzero,omitempty"`
	LinksState       []string            `json:"-" gorm:"type:LowCardinality(String);column:links.span_state" temporaljson:"links_state,omitzero,omitempty"`
	LinksAttributes  []map[string]string `json:"-" gorm:"type:Map(LowCardinality(String), String);column:links.attributes" temporaljson:"links_attributes,omitzero,omitempty"`
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
