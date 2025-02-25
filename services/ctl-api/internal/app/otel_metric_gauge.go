package app

import (
	"time"

	"github.com/nuonco/clickhouse-go/v2"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type OtelMetricGaugeExemplar struct {
	FilteredAttributes map[string]string `json:"filtered_attributes"`
	TimesUnix          string            `json:"times_unix"`
	Value              string            `json:"value"`
	SpanID             string            `json:"span_id"`
	TraceID            string            `json:"trace_id"`
}

type OtelMetricGauge struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"notnull"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	OrgID                  string `json:"org_id"`
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
	Value         float64   `json:"value" gorm:"type:Float64;codec:ZSTD(1)"`
	Flags         uint32    `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)"`

	Exemplars []OtelMetricGaugeExemplar `json:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);"`
}

func (m OtelMetricGauge) GetTableOptions() string {
	opts := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/otel_metrics_gauge', '{replica}')
	TTL toDateTime("time_unix") + toIntervalDay(720)
	PARTITION BY toDate(time_unix)
	PRIMARY KEY (runner_id, runner_job_id, runner_group_id, runner_job_execution_id)
	ORDER BY    (runner_id, runner_job_id, runner_group_id, runner_job_execution_id, toUnixTimestamp64Nano(time_unix), metric_name, attributes)
	SETTINGS index_granularity=8192, ttl_only_drop_parts = 1;`
	return opts
}

func (m OtelMetricGauge) TableName() string {
	return "otel_metrics_gauge"
}

func (r OtelMetricGauge) GetTableClusterOptions() string {
	return "on cluster simple"
}

func (m *OtelMetricGauge) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricGaugeID()
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
type OtelMetricGaugeIngestion struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"not null;default:null"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal attributes
	OrgID                  string `json:"org_id"`
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
	ScopeDroppedAttrCount int               `json:"scope_dropped_attr_count" gorm:"type:UInt32;codec:ZSTD(1)"`
	ScopeSchemaURL        string            `json:"scope_schema_url" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	ServiceName string `json:"service_name" gorm:"type:LowCardinality(String);codec:ZSTD(1)"`

	MetricName        string `json:"metric_name" gorm:"codec:ZSTD(1)"`
	MetricDescription string `json:"metric_description" gorm:"codec:ZSTD(1)"`
	MetricUnit        string `json:"metric_unit" gorm:"codec:ZSTD(1)"`

	Attributes map[string]string `json:"attributes" gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`

	StartTimeUnix time.Time `json:"start_time_unix" gorm:"type:DateTime64(9); codec:Delta, ZSTD(1)"`
	TimeUnix      time.Time `json:"time_unix" gorm:"type:DateTime64(9);codec:ZSTD(1)"`
	Value         float64   `json:"value" gorm:"type:Float64;codec:ZSTD(1)"`
	Flags         uint32    `json:"flags" gorm:"type:UInt32;codec:ZSTD(1)"`

	// Exemplars []OtelMetricGaugeExemplar `json:"-" gorm:"type:Nested(filtered_attributes Map(LowCardinality(String), String), time_unix DateTime64(9), value Float64, span_id String, trace_id String); codec:ZSTD(1);"`
	ExemplarsFilteredAttributes clickhouse.ArraySet `json:"-" gorm:"type:Array(Map(LowCardinality(String), String));column:exemplars.filtered_attributes"`
	ExemplarsTimeUnix           clickhouse.ArraySet `json:"-" gorm:"type:Array(DateTime64(9));column:exemplars.time_unix"`
	ExemplarsValue              clickhouse.ArraySet `json:"-" gorm:"type:Array(Float64);column:exemplars.value"`
	ExemplarsSpanID             clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.span_id"`
	ExemplarsTraceID            clickhouse.ArraySet `json:"-" gorm:"type:Array(String);column:exemplars.trace_id"`
}

func (m *OtelMetricGaugeIngestion) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewOtelMetricGaugeID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if m.OrgID == "" {
		m.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (m OtelMetricGaugeIngestion) TableName() string {
	return "otel_metrics_gauge"
}
