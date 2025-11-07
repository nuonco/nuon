package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallerType string

const (
	InstallerTypeSelfHosted InstallerType = "self_hosted"
)

type Installer struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	Apps []App `json:"apps,omitzero" gorm:"many2many:installer_apps;constraint:OnDelete:CASCADE;" temporaljson:"apps,omitzero,omitempty"`

	Type     InstallerType     `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
	Metadata InstallerMetadata `json:"metadata,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"metadata,omitzero,omitempty"`
}

func (a *Installer) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Installer{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *Installer) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewInstallerID()
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *Installer) AfterQuery(tx *gorm.DB) error {
	return nil
}

func (*Installer) JoinTables() []migrations.JoinTable {
	return []migrations.JoinTable{
		{
			Field:     "Apps",
			JoinTable: &InstallerApp{},
		},
	}
}
