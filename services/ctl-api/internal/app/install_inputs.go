package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallInputs struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`
	OrgID       string                `json:"org_id" gorm:"notnull;default null"`
	Org         Org                   `json:"-" faker:"-"`

	InstallID      string        `json:"install_id" gorm:"notnull;default null"`
	Install        Install       `json:"-"`
	Values         pgtype.Hstore `json:"values"          temporaljson:"values"  gorm:"type:hstore" swaggertype:"object,string"`
	ValuesRedacted pgtype.Hstore `json:"redacted_values" temporaljson:"redacted_values" gorm:"type:hstore;->;-:migration" swaggertype:"object,string"`

	AppInputConfigID string         `json:"app_input_config_id"`
	AppInputConfig   AppInputConfig `json:"-"`
}

func (i *InstallInputs) UseView() bool {
	return true
}

func (i *InstallInputs) ViewVersion() string {
	return "v1"
}

func (i *InstallInputs) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name: views.DefaultViewName(db, &InstallInputs{}, 1),
			SQL:  viewsql.InstallInputsViewV1,
		},
	}
}

func (a *InstallInputs) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	a.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
