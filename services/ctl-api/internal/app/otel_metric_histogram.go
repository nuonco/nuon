package app

import (
	"time"

	"github.com/nuonco/clickhouse-go/v2"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OtelMetricHistogramExemplar struct {
	FilteredAttributes map[string]string `json:"filtered_attributes"`
	TimesUnix          string            `json:"times_unix"`
	Value              string            `json:"value"`
	SpanID             string            `json:"span_id"`
	TraceID            string            `json:"trace_id"`
}

// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_traces.go#L164
type OtelMetricHistogram struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"notnull"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	OrgID                  string `json:"org_id"`
	RunnerID               string `json:"runner_id"`
	RunnerJobID            string `json:"runner_job_id"`
	RunnerGroupID          string `json:"runner_group_id"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step"`

	// OTEL log message attributes
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1)"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ScopeDroppedAttrCount uint32            `json:"scope_dropped_attr_count" gorm:"type:UInt32;codec:ZSTD(1)"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	ServiceName string `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	MetricName        string `json:"metric_name" gorm:"codec:ZSTD(1)"`
	MetricDescription string `json:"metric_description" gorm:"codec:ZSTD(1)"`
	MetricUnit        string `json:"metric_unit" gorm:"codec:ZSTD(1)"`

	Attributes map[string]string `json:"attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`

	StartTimeUnix time.Time `json:"start_time_unix" gorm:"type:DateTime64(9); codec:Delta, ZSTD(1)"`
	TimeUnix      time.Time `json:"time_unix" gorm:"type:DateTime64(9);codec:ZSTD(1)"`

	Count          uint64    `json:"count" gorm:"type:UInt32;codec:ZSTD(1)"`
	Sum            float64   `json:"sum" gorm:"type:Float64;codec:ZSTD(1)"`
	BucketsCount   []uint64  `json:"buckets_count" gorm:"type:Array(UInt64);codec:ZSTD(1)"`
	ExplicitBounds []float64 `json:"explicit_bounds" gorm:"type:Array(Float64);codec:ZSTD(1)"`

	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)"`
	Min   float64 `json:"min" gorm:"type:Float64;codec:ZSTD(1)"`
	Max   float64 `json:"maxx" gorm:"type:Float64;codec:ZSTD(1)"`

	AggregationTemporality int32 `json:"aggregation_temporality" gorm:"type:Int32; codec:ZSTD(1)"`

	Exemplars []OtelMetricHistogramExemplar `json:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);"`
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
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"notnull"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	OrgID                  string `json:"org_id"`
	RunnerID               string `json:"runner_id"`
	RunnerJobID            string `json:"runner_job_id"`
	RunnerGroupID          string `json:"runner_group_id"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step"`

	// OTEL log message attributes
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1)"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ScopeDroppedAttrCount uint32            `json:"scope_dropped_attr_count" gorm:"type:UInt32;codec:ZSTD(1)"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	ServiceName string `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	MetricName        string `json:"metric_name" gorm:"codec:ZSTD(1)"`
	MetricDescription string `json:"metric_description" gorm:"codec:ZSTD(1)"`
	MetricUnit        string `json:"metric_unit" gorm:"codec:ZSTD(1)"`

	Attributes map[string]string `json:"attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`

	StartTimeUnix time.Time `json:"start_time_unix" gorm:"type:DateTime64(9); codec:Delta, ZSTD(1)"`
	TimeUnix      time.Time `json:"time_unix" gorm:"type:DateTime64(9);codec:ZSTD(1)"`

	Count          uint64              `json:"count" gorm:"type:UInt32;codec:ZSTD(1)"`
	Sum            float64             `json:"sum" gorm:"type:Float64;codec:ZSTD(1)"`
	BucketsCount   clickhouse.ArraySet `json:"buckets_count" gorm:"type:Array(UInt64);codec:ZSTD(1)"`
	ExplicitBounds clickhouse.ArraySet `json:"explicit_bounds" gorm:"type:Array(Float64);codec:ZSTD(1)"`

	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)"`
	Min   float64 `json:"min" gorm:"type:Float64;codec:ZSTD(1)"`
	Max   float64 `json:"maxx" gorm:"type:Float64;codec:ZSTD(1)"`

	AggregationTemporality int32 `json:"aggregation_temporality" gorm:"type:Int32; codec:ZSTD(1)"`

	// Exemplars []OtelMetricHistogramExemplar `json:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);"`

	ExemplarsFilteredAttributes clickhouse.ArraySet `json:"-" gorm:"type:Array(Map(LowCardinality(String), String));column:exemplars.filtered_attributes"`
	ExemplarsTimeUnix           clickhouse.ArraySet `json:"-" gorm:"type:Array(DateTime64(9));column:exemplars.time_unix"`
	ExemplarsValue              clickhouse.ArraySet `json:"-" gorm:"type:Array(Float64);column:exemplars.value"`
	ExemplarsSpanID             clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.span_id"`
	ExemplarsTraceID            clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.trace_id"`
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
