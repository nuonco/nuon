package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// clickhouse table
type RunnerHeartBeatV2 struct {
	ID          string `gorm:"primary_key" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id,omitzero" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	RunnerID  string `json:"runner_id,omitzero" temporaljson:"runner_id,omitzero,omitempty"`
	ProcessID string `json:"process_id,omitzero" gorm:"not null;default:''" temporaljson:"process_id,omitzero,omitempty"`

	AliveTime time.Duration `json:"alive_time,omitzero" swaggertype:"primitive,integer" temporaljson:"alive_time,omitzero,omitempty"`
	Version   string        `json:"version,omitzero" temporaljson:"version,omitzero,omitempty"`
	StartedAt time.Time     `json:"started_at,omitzero" gorm:"-" temporaljson:"started_at,omitzero,omitempty"`

	Process RunnerProcessType `json:"process" gorm:"not null;default:''" swaggertype:"string"`
}

func (r *RunnerHeartBeatV2) AfterQuery(tx *gorm.DB) error {
	r.StartedAt = r.CreatedAt.Add(-1 * r.AliveTime)
	return nil
}

func (r *RunnerHeartBeatV2) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerHeartBeatID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r RunnerHeartBeatV2) GetTableOptions() string {
	options := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/runner_heart_beats_v2', '{replica}')
	TTL toDateTime(created_at) + toIntervalDay(2)
	PARTITION BY toDate(created_at)
	PRIMARY KEY (runner_id, process_id, created_at)
	ORDER BY    (runner_id, process_id, created_at)`
	return options
}

func (r RunnerHeartBeatV2) GetTableClusterOptions() string {
	return "on cluster simple"
}

// Struct for a read-only materialized view. the view is created directly in sql.
// NOTE(fd): i am not registering this model so GORM never thinks about it when migrating.
type LatestRunnerHeartBeatV2 struct {
	RunnerID  string            `json:"runner_id,omitzero"  gorm:"->" temporaljson:"runner_id,omitzero,omitempty"`
	ProcessID string            `json:"process_id,omitzero" gorm:"->" temporaljson:"process_id,omitzero,omitempty"`
	Process   RunnerProcessType `json:"process"             gorm:"->" swaggertype:"string"`
	Version   string            `json:"version,omitzero"    gorm:"->" temporaljson:"version,omitzero,omitempty"`
	StartedAt time.Time         `json:"started_at,omitzero" gorm:"-"  temporaljson:"started_at,omitzero,omitempty"`
	AliveTime time.Duration     `json:"alive_time,omitzero" gorm:"->" swaggertype:"primitive,integer" temporaljson:"alive_time,omitzero,omitempty"`
	CreatedAt time.Time         `json:"created_at,omitzero" gorm:"->;column:created_at_latest" temporaljson:"CreatedAt,omitzero,omitempty"`
}

func (r *LatestRunnerHeartBeatV2) AfterQuery(tx *gorm.DB) error {
	r.StartedAt = r.CreatedAt.Add(-1 * r.AliveTime)
	return nil
}

func (*LatestRunnerHeartBeatV2) TableName() string {
	return "latest_runner_heart_beats_v2"
}
