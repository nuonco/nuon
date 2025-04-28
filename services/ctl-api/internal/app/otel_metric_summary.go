package app

import (
	"time"

	"github.com/nuonco/clickhouse-go/v2"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OtelMetricSummaryExemplar struct {
	FilteredAttributes map[string]string `json:"filtered_attributes" temporaljson:"filtered_attributes,omitzero,omitempty"`
	TimesUnix          string            `json:"times_unix" temporaljson:"times_unix,omitzero,omitempty"`
	Value              string            `json:"value" temporaljson:"value,omitzero,omitempty"`
	SpanID             string            `json:"span_id" temporaljson:"span_id,omitzero,omitempty"`
	TraceID            string            `json:"trace_id" temporaljson:"trace_id,omitzero,omitempty"`
}

type OtelMetricSummary struct {
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
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"resource_schema_url,omitzero,omitempty"`
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"resource_attributes,omitzero,omitempty"`

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

	Count uint64  `json:"count" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"count,omitzero,omitempty"`
	Sum   float64 `json:"sum" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"sum,omitzero,omitempty"`
	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"flags,omitzero,omitempty"`

	ValueAtQuantiles map[float64]float64 `json:"value_at_quantiles" gorm:"type:Nested(Quantile Float64,Value Float64);codec:ZSTD(1)" temporaljson:"value_at_quantiles,omitzero,omitempty"`
}

func (m OtelMetricSummary) GetTableOptions() (string, bool) {
	opts := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/otel_metrics_summary', '{replica}')
	TTL toDateTime("time_unix") + toIntervalDay(720)
	PARTITION BY toDate(time_unix)
	PRIMARY KEY (runner_id, runner_job_id, runner_group_id, runner_job_execution_id)
	ORDER BY    (runner_id, runner_job_id, runner_group_id, runner_job_execution_id, toUnixTimestamp64Nano(time_unix), metric_name, attributes)
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
	if m.OrgID == "" {
		m.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

// DO NOT MIGRATE: this is for ingestion only
type OtelMetricSummaryIngestion struct {
	ID          string `gorm:"primary_key" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`

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

	// OTEL attributes
	ResourceAttributes map[string]string `json:"resource_attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"resource_attributes,omitzero,omitempty"`
	ResourceSchemaURL  string            `json:"resource_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"resource_schema_url,omitzero,omitempty"`

	ScopeName             string            `json:"scope_name" gorm:"codec:ZSTD(1)" temporaljson:"scope_name,omitzero,omitempty"`
	ScopeVersion          string            `json:"scope_version" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_version,omitzero,omitempty"`
	ScopeAttributes       map[string]string `json:"scope_attributes" gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"scope_attributes,omitzero,omitempty"`
	ScopeDroppedAttrCount int               `json:"scope_dropped_attr_count" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"scope_dropped_attr_count,omitzero,omitempty"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"scope_schema_url,omitzero,omitempty"`

	ServiceName string `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1)" temporaljson:"service_name,omitzero,omitempty"`

	MetricName        string `json:"metric_name" gorm:"codec:ZSTD(1)" temporaljson:"metric_name,omitzero,omitempty"`
	MetricDescription string `json:"metric_description" gorm:"codec:ZSTD(1)" temporaljson:"metric_description,omitzero,omitempty"`
	MetricUnit        string `json:"metric_unit" gorm:"codec:ZSTD(1)" temporaljson:"metric_unit,omitzero,omitempty"`

	Attributes map[string]string `json:"attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1" temporaljson:"attributes,omitzero,omitempty"`

	StartTimeUnix time.Time `json:"start_time_unix" gorm:"type:DateTime64(9); codec:Delta, ZSTD(1)" temporaljson:"start_time_unix,omitzero,omitempty"`
	TimeUnix      time.Time `json:"time_unix" gorm:"type:DateTime64(9);codec:ZSTD(1)" temporaljson:"time_unix,omitzero,omitempty"`

	Count uint64  `json:"count" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"count,omitzero,omitempty"`
	Sum   float64 `json:"sum" gorm:"type:Float64;codec:ZSTD(1)" temporaljson:"sum,omitzero,omitempty"`
	Flags uint32  `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)" temporaljson:"flags,omitzero,omitempty"`

	// ValueAtQuantiles map[float64]float64 `json:"value_at_quantiles" temporaljson:"value_at_quantiles" gorm:"type:Nested(Quantile Float64,Value Float64);codec:ZSTD(1)"`

	ValueAtQuantilesQuantile clickhouse.ArraySet `json:"-" gorm:"type:Array(Quantile Float64);column:value_at_quantiles.quantile" temporaljson:"value_at_quantiles_quantile,omitzero,omitempty"`
	ValueAtQuantilesValue    clickhouse.ArraySet `json:"-" gorm:"type:Array(Value Float64);column:value_at_quantiles.value" temporaljson:"value_at_quantiles_value,omitzero,omitempty"`
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
