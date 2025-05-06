package app

import (
	"time"

	"github.com/nuonco/clickhouse-go/v2"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OtelMetricHistogramExemplar struct {
	FilteredAttributes map[string]string `json:"filtered_attributes,omitzero" temporaljson:"filtered_attributes,omitzero,omitempty"`
	TimesUnix          string            `json:"times_unix,omitzero" temporaljson:"times_unix,omitzero,omitempty"`
	Value              string            `json:"value,omitzero" temporaljson:"value,omitzero,omitempty"`
	SpanID             string            `json:"span_id,omitzero" temporaljson:"span_id,omitzero,omitempty"`
	TraceID            string            `json:"trace_id,omitzero" temporaljson:"trace_id,omitzero,omitempty"`
}

// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_traces.go#L164
type OtelMetricHistogram struct {
	ID          string `gorm:"primary_key" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// internal attributes
	OrgID                  string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	RunnerID               string `json:"runner_id,omitzero" temporaljson:"runner_id,omitzero,omitempty"`
	RunnerJobID            string `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerGroupID          string `json:"runner_group_id,omitzero" temporaljson:"runner_group_id,omitzero,omitempty"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id,omitzero" temporaljson:"runner_job_execution_id,omitzero,omitempty"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step,omitzero" temporaljson:"runner_job_execution_step,omitzero,omitempty"`

	// OTEL log message attributes
	ResourceAttributes map[string]string `json:"resource_attributes,omitzero" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"resource_attributes,omitzero,omitempty"`
	ResourceSchemaURL  string            `json:"resource_schema_url,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"resource_schema_url,omitzero,omitempty"`

	ScopeName             string            `json:"scope_name,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"scope_name,omitzero,omitempty"`
	ScopeVersion          string            `json:"scope_version,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_version,omitzero,omitempty"`
	ScopeAttributes       map[string]string `json:"scope_attributes,omitzero" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"scope_attributes,omitzero,omitempty"`
	ScopeDroppedAttrCount uint32            `json:"scope_dropped_attr_count,omitzero" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"scope_dropped_attr_count,omitzero,omitempty"`
	ScopeSchemaURL        string            `json:"scope_schema_url,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_schema_url,omitzero,omitempty"`

	ServiceName string `json:"service_name,omitzero" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"service_name,omitzero,omitempty"`

	MetricName        string `json:"metric_name,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"metric_name,omitzero,omitempty"`
	MetricDescription string `json:"metric_description,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"metric_description,omitzero,omitempty"`
	MetricUnit        string `json:"metric_unit,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"metric_unit,omitzero,omitempty"`

	Attributes map[string]string `json:"attributes,omitzero" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"attributes,omitzero,omitempty"`

	StartTimeUnix time.Time `json:"start_time_unix,omitzero" gorm:"type:DateTime64(9); codec:Delta, ZSTD(1)" temporaljson:"start_time_unix,omitzero,omitempty"`
	TimeUnix      time.Time `json:"time_unix,omitzero" gorm:"type:DateTime64(9);codec:ZSTD(1)" temporaljson:"time_unix,omitzero,omitempty"`

	Count          uint64    `json:"count,omitzero" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"count,omitzero,omitempty"`
	Sum            float64   `json:"sum,omitzero" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"sum,omitzero,omitempty"`
	BucketsCount   []uint64  `json:"buckets_count,omitzero" gorm:"type:Array(UInt64);codec:ZSTD(1)" temporaljson:"buckets_count,omitzero,omitempty"`
	ExplicitBounds []float64 `json:"explicit_bounds,omitzero" gorm:"type:Array(Float64);codec:ZSTD(1)" temporaljson:"explicit_bounds,omitzero,omitempty"`

	Flags uint32  `json:"flags,omitzero" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"flags,omitzero,omitempty"`
	Min   float64 `json:"min,omitzero" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"min,omitzero,omitempty"`
	Max   float64 `json:"max,omitzero" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"max,omitzero,omitempty"`

	AggregationTemporality int32 `json:"aggregation_temporality,omitzero" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"aggregation_temporality,omitzero,omitempty"`

	Exemplars []OtelMetricHistogramExemplar `json:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);" temporaljson:"exemplars,omitzero,omitempty"`
}

func (OtelMetricHistogram) GetTableOptions() string {
	opts := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/otel_metrics_histogram', '{replica}')
	TTL toDateTime("time_unix") + toIntervalDay(720)
	PARTITION BY toDate(time_unix)
	PRIMARY KEY (runner_id, runner_job_id, runner_group_id, runner_job_execution_id)
	ORDER BY    (runner_id, runner_job_id, runner_group_id, runner_job_execution_id, toUnixTimestamp64Nano(time_unix), metric_name, attributes)
	SETTINGS index_granularity=8192, ttl_only_drop_parts = 1;`
	return opts
}

func (OtelMetricHistogram) GetTableClusterOptions() string {
	return "on cluster simple"
}

func (m OtelMetricHistogram) TableName() string {
	return "otel_metrics_histogram"
}

func (m *OtelMetricHistogram) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricHistogramID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

// DO NOT MIGRATE
type OtelMetricHistogramIngestion struct {
	ID          string `gorm:"primary_key" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// internal attributes
	OrgID                  string `json:"org_id" temporaljson:"org_id,omitzero,omitempty"`
	RunnerID               string `json:"runner_id" temporaljson:"runner_id,omitzero,omitempty"`
	RunnerJobID            string `json:"runner_job_id" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerGroupID          string `json:"runner_group_id" temporaljson:"runner_group_id,omitzero,omitempty"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id" temporaljson:"runner_job_execution_id,omitzero,omitempty"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step" temporaljson:"runner_job_execution_step,omitzero,omitempty"`

	// OTEL log message attributes
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"resource_attributes,omitzero,omitempty"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"resource_schema_url,omitzero,omitempty"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1)" temporaljson:"scope_name,omitzero,omitempty"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_version,omitzero,omitempty"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"scope_attributes,omitzero,omitempty"`
	ScopeDroppedAttrCount uint32            `json:"scope_dropped_attr_count" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"scope_dropped_attr_count,omitzero,omitempty"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_schema_url,omitzero,omitempty"`

	ServiceName string `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"service_name,omitzero,omitempty"`

	MetricName        string `json:"metric_name" gorm:"codec:ZSTD(1)" temporaljson:"metric_name,omitzero,omitempty"`
	MetricDescription string `json:"metric_description" gorm:"codec:ZSTD(1)" temporaljson:"metric_description,omitzero,omitempty"`
	MetricUnit        string `json:"metric_unit" gorm:"codec:ZSTD(1)" temporaljson:"metric_unit,omitzero,omitempty"`

	Attributes map[string]string `json:"attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"attributes,omitzero,omitempty"`

	StartTimeUnix time.Time `json:"start_time_unix" gorm:"type:DateTime64(9); codec:Delta, ZSTD(1)" temporaljson:"start_time_unix,omitzero,omitempty"`
	TimeUnix      time.Time `json:"time_unix" gorm:"type:DateTime64(9);codec:ZSTD(1)" temporaljson:"time_unix,omitzero,omitempty"`

	Count          uint64              `json:"count" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"count,omitzero,omitempty"`
	Sum            float64             `json:"sum" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"sum,omitzero,omitempty"`
	BucketsCount   clickhouse.ArraySet `json:"buckets_count" gorm:"type:Array(UInt64);codec:ZSTD(1)" temporaljson:"buckets_count,omitzero,omitempty"`
	ExplicitBounds clickhouse.ArraySet `json:"explicit_bounds" gorm:"type:Array(Float64);codec:ZSTD(1)" temporaljson:"explicit_bounds,omitzero,omitempty"`

	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"flags,omitzero,omitempty"`
	Min   float64 `json:"min" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"min,omitzero,omitempty"`
	Max   float64 `json:"maxx" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"max,omitzero,omitempty"`

	AggregationTemporality int32 `json:"aggregation_temporality" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"aggregation_temporality,omitzero,omitempty"`

	// Exemplars []OtelMetricHistogramExemplar `json:"-" temporaljson:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);"`

	ExemplarsFilteredAttributes clickhouse.ArraySet `json:"-" gorm:"type:Array(Map(LowCardinality(String), String));column:exemplars.filtered_attributes" temporaljson:"exemplars_filtered_attributes,omitzero,omitempty"`
	ExemplarsTimeUnix           clickhouse.ArraySet `json:"-" gorm:"type:Array(DateTime64(9));column:exemplars.time_unix" temporaljson:"exemplars_time_unix,omitzero,omitempty"`
	ExemplarsValue              clickhouse.ArraySet `json:"-" gorm:"type:Array(Float64);column:exemplars.value" temporaljson:"exemplars_value,omitzero,omitempty"`
	ExemplarsSpanID             clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.span_id" temporaljson:"exemplars_span_id,omitzero,omitempty"`
	ExemplarsTraceID            clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.trace_id" temporaljson:"exemplars_trace_id,omitzero,omitempty"`
}

func (m OtelMetricHistogramIngestion) TableName() string {
	return "otel_metrics_histogram"
}

func (m *OtelMetricHistogramIngestion) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricHistogramID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
