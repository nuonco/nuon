package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
)

type RunnerGroupType string

const (
	RunnerGroupTypeInstall RunnerGroupType = "install"
	RunnerGroupTypeOrg     RunnerGroupType = "org"
)

type RunnerGroup struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `gorm:"index:idx_runner_group_owner" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"default null;not null" temporaljson:"org_id,omitzero,omitempty"`

	// parent can org, install or in the future, builtin runner group
	OwnerID   string `json:"owner_id,omitzero" gorm:"index:idx_runner_group_owner;notnull;default null" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"notnull;default null" temporaljson:"owner_type,omitzero,omitempty"`

	Runners  []Runner            `json:"runners,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"runners,omitzero,omitempty"`
	Settings RunnerGroupSettings `json:"settings,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"settings,omitzero,omitempty"`
	Type     RunnerGroupType     `json:"type,omitzero" gorm:"notnull;defaultnull" temporaljson:"type,omitzero,omitempty"`
	Platform AppRunnerType       `json:"platform,omitzero" gorm:"notnull;defaultnull" temporaljson:"platform,omitzero,omitempty"`
}

func (r *RunnerGroup) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerGroup{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *RunnerGroup) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerGroupID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r *RunnerGroup) EventLoops() []bulk.EventLoop {
	evs := make([]bulk.EventLoop, 0)
	for _, runner := range r.Runners {
		evs = append(evs, bulk.EventLoop{
			Namespace: "runners",
			ID:        runner.ID,
		})
	}

	return evs
}
