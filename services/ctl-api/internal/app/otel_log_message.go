package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

const OtelLogTableOptions string = `ENGINE = MergeTree()
	PARTITION BY toDate(timestamp_time)
	PRIMARY KEY (service_name, timestamp_time)
	ORDER BY (service_name, timestamp_time, timestamp)
	TTL timestamp_time + toIntervalDay(180)
	SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;
`

// Logs are designed to be written via an OTLP exporter.
//
// https://opentelemetry.io/docs/specs/otel/logs/bridge-api/
//
// The clickhouse exporter, is a good reference point for this
// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_logs.go
type OtelLogRecord struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"notnull"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal log attributes
	RunnerID             string `json:"runner_id"`
	RunnerJobID          string `json:"runner_job_id"`
	RunnerJobExecutionID string `json:"runner_job_execution_id"`

	// OTEL log message attributes
	Timestamp          time.Time         `gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1)"`
	TimestampDate      time.Time         `gorm:"type:Date DEFAULT toDate(timestamp)"`
	TimestampTime      time.Time         `gorm:"type:DateTime DEFAULT toDateTime(timestamp)"`
	TraceID            string            `gorm:"codec:ZSTD(1);index:idx_trace_id,type:bloom_filter(0.001),granularity:1;"`
	SpanID             string            `gorm:"codec:ZSTD(1)"`
	TraceFlags         int               `gorm:"type:UInt8"`
	SeverityText       string            `gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	SeverityNumber     int               `gorm:"type:UInt8"`
	ServiceName        string            `gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	Body               string            `gorm:"codecZSTD(1);index:idx_body,type:tokenbf_v1(32768\\,3\\,0),granularity:8;"`
	ResourceSchemaURL  string            `gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ResourceAttributes map[string]string `gorm:"type:Map(LowCardinality(String),String);codec:ZSTD(1); index:idx_res_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_res_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	ScopeSchemaURL     string            `gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ScopeName          string            `gorm:"codec:ZSTD(1)"`
	ScopeVersion       string            `gorm:"type:LowCardinality(String);codec:ZSTD(1)"`
	ScopeAttributes    map[string]string `gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1);index:idx_scope_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_scope_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
	LogAttributes      map[string]string `gorm:"type:Map(LowCardinality(String), String);codec:ZSTD(1); index:idx_log_attr_key,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1; index:idx_log_attr_value,expression:mapKeys(resource_attributes),type:bloom_filter(0.1),granularity:1"`
}

func (r *OtelLogRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewOtelLogID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (r OtelLogRecord) GetTableOptions() (string, bool) {
	opts := `ENGINE = MergeTree()
	PARTITION BY toDate(timestamp_time)
	PRIMARY KEY (service_name, timestamp_time)
	ORDER BY (service_name, timestamp_time, timestamp)
	TTL timestamp_time + toIntervalDay(180)
	SETTINGS index_granularity = 8192, ttl_only_drop_parts = 1;`
	return opts, true
}

func (r OtelLogRecord) MigrateDB(db *gorm.DB) *gorm.DB {
	opts, hasOpts := r.GetTableOptions()
	if !hasOpts {
		return db
	}
	return db.Set("gorm:table_options", opts)
}
