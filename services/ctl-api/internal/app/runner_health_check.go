package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

// clickhouse table
type RunnerHealthCheck struct {
	ID          string `gorm:"primary_key" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id,omitzero" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1)" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"type:DateTime64(9);codec:Delta(8),ZSTD(1)" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	RunnerID     string       `json:"runner_id,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"runner_id,omitzero,omitempty"`
	RunnerJob    RunnerJob    `json:"runner_job,omitzero" gorm:"polymorphic:Owner;" temporaljson:"runner_job,omitzero,omitempty"`
	RunnerStatus RunnerStatus `json:"status,omitzero" gorm:"codec:ZSTD(1)" temporaljson:"runner_status,omitzero,omitempty"`

	// loaded from view

	MinuteBucket time.Time `json:"minute_bucket,omitzero" gorm:"->;-:migration;type:DateTime64(9);codec:Delta(8),ZSTD(1)" temporaljson:"minute_bucket,omitzero,omitempty"`

	// after queries

	RunnerStatusCode int `json:"status_code" gorm:"-" temporaljson:"runner_status_code,omitzero,omitempty"`

	Process RunnerHeartBeatProcess `json:"process" gorm:"not null;default:install" swaggertype:"string"`
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

func (r RunnerHealthCheck) GetTableOptions() string {
	options := `ENGINE = ReplicatedMergeTree('/var/lib/clickhouse/{cluster}/tables/{shard}/{uuid}/runner_health_checks', '{replica}')
	TTL toDateTime(created_at) + toIntervalDay(1)
	PARTITION BY toDate(created_at)
	PRIMARY KEY (runner_id, created_at)
	ORDER BY    (runner_id, created_at)`
	return options
}

func (r RunnerHealthCheck) GetTableClusterOptions() string {
	return "on cluster simple"
}

func (*RunnerHealthCheck) UseView() bool {
	return false
}

func (*RunnerHealthCheck) ViewVersion() string {
	return "v1"
}

func (i *RunnerHealthCheck) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &RunnerHealthCheck{}, 1),
			SQL:  viewsql.RunnerHealthCheckViewV1,
		},
	}
}

func (r *RunnerHealthCheck) AfterQuery(tx *gorm.DB) error {
	r.RunnerStatusCode = r.RunnerStatus.Code()
	return nil
}
