package app

import (
	"time"

	"github.com/nuonco/clickhouse-go/v2"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OtelMetricExponentialHistogramExemplar struct {
	FilteredAttributes map[string]string `json:"filtered_attributes" temporaljson:"filtered_attributes,omitzero,omitempty"`
	TimesUnix          string            `json:"times_unix" temporaljson:"times_unix,omitzero,omitempty"`
	Value              string            `json:"value" temporaljson:"value,omitzero,omitempty"`
	SpanID             string            `json:"span_id" temporaljson:"span_id,omitzero,omitempty"`
	TraceID            string            `json:"trace_id" temporaljson:"trace_id,omitzero,omitempty"`
}

// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_traces.go#L164
type OtelMetricExponentialHistogram struct {
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

	Count                uint64   `json:"count" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"count,omitzero,omitempty"`
	Sum                  float64  `json:"sum" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"sum,omitzero,omitempty"`
	Scale                float64  `json:"scale" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"scale,omitzero,omitempty"`
	ZeroCount            uint64   `json:"zero_count" gorm:"type:UInt64;codec:ZSTD(1)" temporaljson:"zero_count,omitzero,omitempty"`
	PositiveOffset       int32    `json:"positive_offset" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"positive_offset,omitzero,omitempty"`
	PositiveBucketCounts []uint64 `json:"positive_bucket_counts" gorm:"type:Array(UInt64); codec:ZSTD(1)" temporaljson:"positive_bucket_counts,omitzero,omitempty"`
	NegativeOffset       int32    `json:"negative_offset" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"negative_offset,omitzero,omitempty"`
	NegativeBucketCounts []uint64 `json:"negative_bucket_counts" gorm:"type:Array(UInt64); codec:ZSTD(1)" temporaljson:"negative_bucket_counts,omitzero,omitempty"`

	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"flags,omitzero,omitempty"`
	Min   float64 `json:"min" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"min,omitzero,omitempty"`
	Max   float64 `json:"maxx" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"max,omitzero,omitempty"`

	AggregationTemporality int32 `json:"aggregation_temporality" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"aggregation_temporality,omitzero,omitempty"`

	Exemplars []OtelMetricExponentialHistogramExemplar `json:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);" temporaljson:"exemplars,omitzero,omitempty"`
}

func (m OtelMetricExponentialHistogram) GetTableOptions() (string, bool) {
	opts := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/otel_metrics_exponential_histogram', '{replica}')
	TTL toDateTime("time_unix") + toIntervalDay(720)
	PARTITION BY toDate(time_unix)
	PRIMARY KEY (runner_id, runner_job_id, runner_group_id, runner_job_execution_id)
	ORDER BY    (runner_id, runner_job_id, runner_group_id, runner_job_execution_id, toUnixTimestamp64Nano(time_unix), metric_name, attributes)
	SETTINGS index_granularity=8192, ttl_only_drop_parts = 1;`
	return opts, true
}

func (m OtelMetricExponentialHistogram) TableName() string {
	return "otel_metrics_exponential_histogram"
}

func (m *OtelMetricExponentialHistogram) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricExponentialHistogramID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if m.OrgID == "" {
		m.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

// DO NOT MIGRATE
type OtelMetricExponentialHistogramIngestion struct {
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

	Count                uint64              `json:"count" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"count,omitzero,omitempty"`
	Sum                  float64             `json:"sum" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"sum,omitzero,omitempty"`
	Scale                int32               `json:"scale" gorm:"type:Int32;codec:ZSTD(1)" temporaljson:"scale,omitzero,omitempty"`
	ZeroCount            uint64              `json:"zero_count" gorm:"type:UInt64;codec:ZSTD(1)" temporaljson:"zero_count,omitzero,omitempty"`
	PositiveOffset       int32               `json:"positive_offset" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"positive_offset,omitzero,omitempty"`
	PositiveBucketCounts clickhouse.ArraySet `json:"positive_bucket_counts" gorm:"type:Array(UInt64); codec:ZSTD(1)" temporaljson:"positive_bucket_counts,omitzero,omitempty"`
	NegativeOffset       int32               `json:"negative_offset" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"negative_offset,omitzero,omitempty"`
	NegativeBucketCounts clickhouse.ArraySet `json:"negative_bucket_counts" gorm:"type:Array(UInt64); codec:ZSTD(1)" temporaljson:"negative_bucket_counts,omitzero,omitempty"`

	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"flags,omitzero,omitempty"`
	Min   float64 `json:"min" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"min,omitzero,omitempty"`
	Max   float64 `json:"maxx" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"max,omitzero,omitempty"`

	AggregationTemporality int32 `json:"aggregation_temporality" gorm:"type:Int32; codec:ZSTD(1)" temporaljson:"aggregation_temporality,omitzero,omitempty"`

	// Exemplars []OtelMetricExponentialHistogramExemplar `json:"-" temporaljson:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);"`

	ExemplarsFilteredAttributes clickhouse.ArraySet `json:"-" gorm:"type:Array(Map(LowCardinality(String), String));column:exemplars.filtered_attributes" temporaljson:"exemplars_filtered_attributes,omitzero,omitempty"`
	ExemplarsTimeUnix           clickhouse.ArraySet `json:"-" gorm:"type:Array(DateTime64(9));column:exemplars.time_unix" temporaljson:"exemplars_time_unix,omitzero,omitempty"`
	ExemplarsValue              clickhouse.ArraySet `json:"-" gorm:"type:Array(Float64);column:exemplars.value" temporaljson:"exemplars_value,omitzero,omitempty"`
	ExemplarsSpanID             clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.span_id" temporaljson:"exemplars_span_id,omitzero,omitempty"`
	ExemplarsTraceID            clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.trace_id" temporaljson:"exemplars_trace_id,omitzero,omitempty"`
}

func (m OtelMetricExponentialHistogramIngestion) TableName() string {
	return "otel_metrics_exponential_histogram"
}

func (m *OtelMetricExponentialHistogramIngestion) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricExponentialHistogramID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if m.OrgID == "" {
		m.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
