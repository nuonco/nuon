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

	RunnerStatus RunnerStatus `json:"status"`

	// after queries

	RunnerStatusCode int       `json:"status_code" gorm:"-"`
	MinuteBucket     time.Time `json:"minute_bucket" gorm:"-"`
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
	options := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/runner_health_checks', '{replica}')
	PARTITION BY toDate(created_at)
	PRIMARY KEY (runner_id, created_at)
	ORDER BY    (runner_id, created_at)`
	return options, true
}

func (r RunnerHealthCheck) MigrateDB(tx *gorm.DB) *gorm.DB {
	opts, hasOpts := r.GetTableOptions()
	if !hasOpts {
		return tx
	}
	return tx.Set("gorm:table_options", opts).Set("gorm:table_cluster_options", "on cluster simple")
}

func (r *RunnerHealthCheck) AfterQuery(tx *gorm.DB) error {
	r.RunnerStatusCode = r.RunnerStatus.Code()
	r.MinuteBucket = r.CreatedAt.Round(time.Minute)
	return nil
}
