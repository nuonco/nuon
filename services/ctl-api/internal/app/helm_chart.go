package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type HelmChart struct {
	ID          string                `gorm:"primary_key:true;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" `
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" `
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" `
	DeletedAt   soft_delete.DeletedAt `json:"-" `

	OrgID string `json:"org_id" `
	Org   Org    `json:"-" `

	OwnerID   string `json:"owner_id" gorm:"type:text;check:owner_id_checker,char_length(id)=26;uniqueIndex:idx_owner" `
	OwnerType string `json:"owner_type" gorm:"type:text;uniqueIndex:idx_owner"`

	HelmReleases []HelmRelease `faker:"-"  gorm:"constraint:OnDelete:CASCADE;"`
}

func (h *HelmChart) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &HelmChart{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (h *HelmChart) BeforeCreate(tx *gorm.DB) (err error) {
	if h.ID == "" {
		h.ID = domains.NewHelmChartID()
	}

	if h.CreatedByID == "" {
		h.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if h.OrgID == "" {
		h.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil

}
