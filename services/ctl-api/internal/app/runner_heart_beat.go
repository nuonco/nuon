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
	Version   string        `json:"version"`
	StartedAt time.Time     `json:"started_at" gorm:"-"`
}

func (r *RunnerHeartBeat) AfterQuery(tx *gorm.DB) error {
	r.StartedAt = r.CreatedAt.Add(-1 * r.AliveTime)
	return nil
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

func (r RunnerHeartBeat) GetTableOptions() string {
	options := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/runner_heart_beats', '{replica}')
	TTL toDateTime(created_at) + toIntervalDay(3)
	PARTITION BY toDate(created_at)
	PRIMARY KEY (runner_id, created_at)
	ORDER BY    (runner_id, created_at)`
	return options
}

func (r RunnerHeartBeat) GetTableClusterOptions() string {
	return "on cluster simple"
}
