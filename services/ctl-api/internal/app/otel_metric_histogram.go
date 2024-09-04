package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/clickhouseexporter/exporter_traces.go#L164
type OtelMetricHistogram struct {
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
	TraceID            string
	SpanID             string
	ParentSpanID       string
	TraceState         string
	SpanName           string
	SpanKind           string
	ServiceName        string
	ResourceAttributes string
	ScopeName          string
	ScopeVersion       string
	SpanAttributes     string
	Duration           int64
	StatusCode         string
	StatusMessage      string

	// TODO(jm): add events
	Events string

	// TODO(jm): add links
	Links string
}
