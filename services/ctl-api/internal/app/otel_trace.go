package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type OtelTrace struct {
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
	ResourceAttributes     map[string]string
	ResourceSchemaURL      string
	ScopeName              string
	ScopeVersion           string
	ScopeAttributes        map[string]string
	ScopeDroppedAttrCount  int
	ScopeSchemaURL         string
	ServiceName            string
	MetricName             string
	MetricDescription      string
	MetricUnit             string
	Attributes             map[string]string
	StartTimeUnix          time.Time
	TimeUnix               time.Time
	Count                  int64
	Sum                    float64
	BucketCounts           []int
	ExplicitBounds         []float64
	Flags                  int
	Min                    float64
	Max                    float64
	AggregationTemporality int
}
