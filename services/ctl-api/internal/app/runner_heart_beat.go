package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type RunnerHeartBeatProcess string

const (
	RunnerHeartBeatProcessMng     RunnerHeartBeatProcess = "mng"
	RunnerHeartBeatProcessInstall RunnerHeartBeatProcess = "install"
	RunnerHeartBeatProcessOrg     RunnerHeartBeatProcess = "org"
)

// clickhouse table
type RunnerHeartBeat struct {
	ID          string `gorm:"primary_key" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id,omitzero" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	RunnerID string `json:"runner_id,omitzero" temporaljson:"runner_id,omitzero,omitempty"`

	AliveTime time.Duration `json:"alive_time,omitzero" swaggertype:"primitive,integer" temporaljson:"alive_time,omitzero,omitempty"`
	Version   string        `json:"version,omitzero" temporaljson:"version,omitzero,omitempty"`
	StartedAt time.Time     `json:"started_at,omitzero" gorm:"-" temporaljson:"started_at,omitzero,omitempty"`

	Process RunnerHeartBeatProcess `json:"process" gorm:"not null;default:install" swaggertype:"string"`
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
