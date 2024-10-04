package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

// clickhouse table
type RunnerHealthCheck struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	RunnerID string `json:"runner_id"`
}

func (r *RunnerHealthCheck) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerHealthCheckID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r RunnerHealthCheck) GetTableOptions() (string, bool) {
	options := `ENGINE = ReplicatedMergeTree('/clickhouse/{cluster}/tables/{shard}/runner_health_checks', '{replica}')
	ORDER BY (created_at)
	PARTITION BY toDate(created_at)
	PRIMARY KEY (runner_id, created_at)`
	return options, true
}

func (r RunnerHealthCheck) MigrateDB(tx *gorm.DB) *gorm.DB {
	opts, hasOpts := r.GetTableOptions()
	if !hasOpts {

		return tx
	}
	return tx.Set("gorm:table_options", opts)
}
