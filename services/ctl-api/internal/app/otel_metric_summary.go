package app

import (
	"time"

	"github.com/nuonco/clickhouse-go/v2"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OtelMetricSummaryExemplar struct {
	FilteredAttributes map[string]string `json:"filtered_attributes"`
	TimesUnix          string            `json:"times_unix"`
	Value              string            `json:"value"`
	SpanID             string            `json:"span_id"`
	TraceID            string            `json:"trace_id"`
}

type OtelMetricSummary struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"notnull"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	RunnerID               string `json:"runner_id" `
	RunnerJobID            string `json:"runner_job_id"`
	RunnerGroupID          string `json:"runner_group_id"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step"`

	// OTEL log message attributes
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`

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

	Count uint64  `json:"count" gorm:"type:UInt32;codec:ZSTD(1)"`
	Sum   float64 `json:"sum" gorm:"type:Float64;codec:ZSTD(1)"`
	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)"`

	ValueAtQuantiles map[float64]float64 `json:"value_at_quantiles" gorm:"type:Nested(Quantile Float64,Value Float64);codec:ZSTD(1)"`
}

func (m OtelMetricSummary) GetTableOptions() (string, bool) {
	opts := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/otel_metrics_summary', '{replica}')
	TTL toDateTime("time_unix") + toIntervalDay(720)
	PARTITION BY toDate(time_unix)
	ORDER BY (service_name, metric_name, attributes, toUnixTimestamp64Nano(time_unix))
	SETTINGS index_granularity=8192, ttl_only_drop_parts = 1;`
	return opts, true
}

func (m OtelMetricSummary) TableName() string {
	return "otel_metrics_summary"
}

func (m OtelMetricSummary) MigrateDB(db *gorm.DB) *gorm.DB {
	opts, hasOpts := m.GetTableOptions()
	if !hasOpts {
		return db
	}
	return db.Set("gorm:table_options", opts).Set("gorm:table_cluster_options", "on cluster simple")
}

func (m *OtelMetricSummary) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricSummaryID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

// DO NOT MIGRATE: this is for ingestion only
type OtelMetricSummaryIngestion struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"not null;default:null"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	RunnerID               string `json:"runner_id" `
	RunnerJobID            string `json:"runner_job_id"`
	RunnerGroupID          string `json:"runner_group_id"`
	RunnerJobExecutionID   string `json:"runner_job_execution_id"`
	RunnerJobExecutionStep string `json:"runner_job_execution_step"`

	// OTEL attributes
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1)"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ScopeDroppedAttrCount int               `json:"scope_dropped_attr_count" gorm:"type:UInt32;codec:ZSTD(1)"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	ServiceName string `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	MetricName        string `json:"metric_name" gorm:"codec:ZSTD(1)"`
	MetricDescription string `json:"metric_description" gorm:"codec:ZSTD(1)"`
	MetricUnit        string `json:"metric_unit" gorm:"codec:ZSTD(1)"`

	Attributes map[string]string `json:"attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`

	StartTimeUnix time.Time `json:"start_time_unix" gorm:"type:DateTime64(9); codec:Delta, ZSTD(1)"`
	TimeUnix      time.Time `json:"time_unix" gorm:"type:DateTime64(9);codec:ZSTD(1)"`

	Count uint64  `json:"count" gorm:"type:UInt32;codec:ZSTD(1)"`
	Sum   float64 `json:"sum" gorm:"type:Float64;codec:ZSTD(1)"`
	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)"`

	// ValueAtQuantiles map[float64]float64 `json:"value_at_quantiles" gorm:"type:Nested(Quantile Float64,Value Float64);codec:ZSTD(1)"`

	ValueAtQuantilesQuantile clickhouse.ArraySet `json:"-" gorm:"type:Array(Quantile Float64);column:value_at_quantiles.quantile"`
	ValueAtQuantilesValue    clickhouse.ArraySet `json:"-" gorm:"type:Array(Value Float64);column:value_at_quantiles.value"`
}

func (m *OtelMetricSummaryIngestion) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricSummaryID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (m OtelMetricSummaryIngestion) TableName() string {
	return "otel_metrics_summmary"
}
