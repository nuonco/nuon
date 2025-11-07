package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type AppSecret struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_secret_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID string `json:"app_id,omitzero" gorm:"not null;default null;index:idx_app_secret_name,unique" temporaljson:"app_id,omitzero,omitempty"`
	App   App    `json:"-" faker:"-" temporaljson:"app,omitzero,omitempty"`

	Name  string `json:"name,omitzero" gorm:"not null;default null;index:idx_app_secret_name,unique" temporaljson:"name,omitzero,omitempty"`
	Value string `json:"-" gorm:"not null;default null" temporaljson:"value,omitzero,omitempty"`

	// after query fields
	Length int `json:"length,omitzero" gorm:"-" temporaljson:"length,omitzero,omitempty"`
}

func (a *AppSecret) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppSecret{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *AppSecret) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppSecretID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *AppSecret) AfterQuery(tx *gorm.DB) error {
	a.Length = len(a.Value)
	return nil
}
