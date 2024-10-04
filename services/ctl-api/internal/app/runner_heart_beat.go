package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

// clickhouse table
type RunnerHeartBeat struct {
	ID          string `gorm:"primary_key" json:"id"`
	CreatedByID string `json:"created_by_id"`

	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	RunnerID string `json:"runner_id"`

	AliveTime time.Duration `json:"alive_time" swaggertype:"primitive,integer"`
}

func (r *RunnerHeartBeat) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerHeartBeatID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r RunnerHeartBeat) GetTableOptions() (string, bool) {
	options := `ENGINE = ReplicatedMergeTree('/clickhouse/{cluster}/tables/{shard}/runner_heart_beats', '{replica}')
	ORDER BY (created_at, runner_id)
	PARTITION BY toDate(created_at)
	PRIMARY KEY (created_at, runner_id)`
	return options, true
}

func (r RunnerHeartBeat) MigrateDB(tx *gorm.DB) *gorm.DB {
	opts, hasOpts := r.GetTableOptions()
	if !hasOpts {
		return tx
	}
	return tx.Set("gorm:table_options", opts).Set("gorm:table_cluster_options", "on cluster simple")
}
