package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type TerraformWorkspace struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	OwnerID   string `json:"owner_id,omitzero" gorm:"type:text;check:owner_id_checker,char_length(id)=26;uniqueIndex:idx_owner" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type,omitzero" gorm:"type:text;uniqueIndex:idx_owner" temporaljson:"owner_type,omitzero,omitempty"`

	States      []TerraformWorkspaceState `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"states,omitzero,omitempty"`
	LockHistory []TerraformWorkspaceLock  `faker:"-" json:"lock_history" swaggerignore:"true" gorm:"foreignKey:WorkspaceID;references:ID;constraint:OnDelete:CASCADE;" temporaljson:"lock_history,omitzero,omitempty"`

	StateJSON []TerraformWorkspaceStateJSON `json:"-" swaggerignore:"true" gorm:"foreignKey:WorkspaceID;references:ID;constraint:OnDelete:CASCADE;" temporaljson:"statesjson,omitzero,omitempty"`
}

func (r *TerraformWorkspace) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &TerraformWorkspace{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *TerraformWorkspace) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == "" {
		r.ID = domains.NewTerraformWorkspaceID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
