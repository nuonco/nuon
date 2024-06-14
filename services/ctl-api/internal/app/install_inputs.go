package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallInputs struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`
	OrgID       string                `json:"org_id" gorm:"notnull;default null"`
	Org         Org                   `json:"-" faker:"-"`

	InstallID string        `json:"install_id" gorm:"notnull;default null"`
	Install   Install       `json:"-"`
	Values    pgtype.Hstore `json:"values" gorm:"type:hstore" swaggertype:"object,string"`

	AppInputConfigID string         `json:"app_input_config_id"`
	AppInputConfig   AppInputConfig `json:"-"`
}

func (a *InstallInputs) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
