package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallPlan struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string `json:"install_id,omitzero" gorm:"defaultnull;notnull;" temporaljson:"install_id,omitzero,omitempty"`

	OwnerID     string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType   string `json:"owner_type,omitzero" gorm:"type:text;" temporaljson:"owner_type,omitzero,omitempty"`
	PlanJSON    string `json:"plan_json,omitzero" temporaljson:"plan_json,omitzero,omitempty"`       // NOTE: a bit of a mis-nomer. atm, this stores b64-encoded bytes
	PlanDisplay string `json:"plan_display,omitzero" temporaljson:"plan_display,omitzero,omitempty"` // NOTE: this stores a human legible plan
}

func (r *InstallPlan) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
