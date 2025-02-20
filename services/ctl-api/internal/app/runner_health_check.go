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

	CreatedAt time.Time             `json:"created_at" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1)"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1)"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	RunnerID     string       `json:"runner_id" gorm:"codec:ZSTD(1)"`
	RunnerJob    RunnerJob    `json:"runner_job" gorm:"polymorphic:Owner;"`
	RunnerStatus RunnerStatus `json:"status" gorm:"codec:ZSTD(1)"`

	// loaded from view

	MinuteBucket time.Time `json:"minute_bucket" gorm:"->;-:migration;type:DateTime64(9);codec:Delta(8),ZSTD(1)"`

	// after queries

	RunnerStatusCode int `json:"status_code" gorm:"-"`
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
	TTL toDateTime(created_at) + toIntervalDay(7)
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

func (*RunnerHealthCheck) UseView() bool {
	return true
}

func (*RunnerHealthCheck) ViewVersion() string {
	return "v1"
}

func (r *RunnerHealthCheck) AfterQuery(tx *gorm.DB) error {
	r.RunnerStatusCode = r.RunnerStatus.Code()
	return nil
}
