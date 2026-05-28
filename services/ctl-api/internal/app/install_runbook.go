package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallRunbook struct {
	ID          string                `json:"id,omitzero" gorm:"primary_key;check:id_checker,char_length(id)=26" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_install_runbook_id,unique;index:idx_irb_org_id_install_id,priority:3" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;index:idx_irb_org_id_install_id,priority:1" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Install   Install `json:"-" swaggerignore:"true" temporaljson:"install,omitzero,omitempty"`
	InstallID string  `json:"install_id,omitzero" gorm:"index:idx_install_runbook_id,unique;index:idx_irb_org_id_install_id,priority:2" faker:"-" temporaljson:"install_id,omitzero,omitempty"`

	Runbook   Runbook `json:"runbook,omitzero" temporaljson:"runbook,omitzero,omitempty"`
	RunbookID string  `json:"runbook_id,omitzero" gorm:"index:idx_install_runbook_id,unique" temporaljson:"runbook_id,omitzero,omitempty"`

	Runs []InstallRunbookRun `faker:"-" gorm:"constraint:OnDelete:CASCADE;" json:"runs,omitzero" temporaljson:"runs,omitzero,omitempty"`

	// after query fields
	Status InstallRunbookRunStatus `json:"status,omitzero" gorm:"-" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
}

func (r *InstallRunbook) BeforeCreate(tx *gorm.DB) error {
	r.ID = domains.NewInstallRunbookID()
	r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	r.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (r *InstallRunbook) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallRunbook{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *InstallRunbook) AfterQuery(tx *gorm.DB) error {
	r.Status = InstallRunbookRunStatusUnknown
	if len(r.Runs) > 0 {
		r.Status = r.Runs[0].Status
	}
	return nil
}
