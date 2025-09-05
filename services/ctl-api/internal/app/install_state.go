package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallState struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Install   Install `json:"-" faker:"-" temporaljson:"install,omitzero,omitempty"`
	InstallID string  `json:"install_id,omitzero" gorm:"notnull" temporaljson:"install_id,omitzero,omitempty"`

	State   *state.State `json:"contents,omitzero" gorm:"type:jsonb" swaggertype:"string" temporaljson:"-"`
	Version int          `json:"version,omitzero" gorm:"->;-:migration" temporaljson:"version,omitzero,omitempty"`

	TriggeredByID   string `json:"triggered_by_id,omitzero" gorm:"type:text;check:triggered_by_id_checker,char_length(id)=26"`
	TriggeredByType string `json:"triggered_by_type,omitzero" gorm:"type:text;"`

	Archived bool `json:"archived" gorm:"default:false;not null" temporaljson:"archived,omitzero,omitempty"`

	StaleAt generics.NullTime `json:"stale_at,omitzero" gorm:"type:timestamp;default:null" temporaljson:"stale_at,omitzero,omitempty"`
}

func (a *InstallState) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewInstallStateID()
	}

	// NOTE: temporary but we need to fallback to the install's created_by_id and org_id if not set
	var install *Install
	if a.CreatedByID == "" || a.OrgID == "" {
		res := tx.WithContext(tx.Statement.Context).
			Model(&Install{}).
			First(&install, "id = ?", a.InstallID)
		if res.Error != nil {
			return res.Error
		}
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
		if a.CreatedByID == "" {
			if install != nil {
				a.CreatedByID = install.CreatedByID
			} else {
				return fmt.Errorf("created_by_id is required and could not be determined")
			}
		}
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
		if a.OrgID == "" {
			if install != nil {
				a.OrgID = install.OrgID
			} else {
				return fmt.Errorf("org_id is required and could not be determined")
			}
		}
	}

	return nil
}

func (i *InstallState) UseView() bool {
	return true
}

func (i *InstallState) ViewVersion() string {
	return "v1"
}

func (i *InstallState) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &InstallState{}, 1),
			SQL:           viewsql.InstallStatesViewV1,
			AlwaysReapply: true,
		},
	}
}
