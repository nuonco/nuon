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
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id" gorm:"not null;default:null"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	// internal log attributes
	RunnerID             string `json:"runner_id" `
	RunnerJobID          string `json:"runner_job_id"`
	RunnerJobExecutionID string `json:"runner_job_execution_id"`

	// OTEL log message attributes
	Timestamp          time.Time
	TimestampDate      time.Time
	TimestampTime      time.Time
	TraceID            string
	SpanID             string
	TraceFlags         int
	SeverityText       string
	SeverityNumber     int
	ServiceName        string
	Body               string
	ResourceSchemaURL  string
	ResourceAttributes map[string]string
	ScopeSchemaURL     string
	ScopeName          string
	ScopeVersion       string
	ScopeAttributes    map[string]string
	LogAttributes      map[string]string
}

func (r *OtelLogRecord) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}
