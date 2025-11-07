package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallInputs struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	OrgID       string                `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org         Org                   `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID      string        `json:"install_id,omitzero" gorm:"notnull;default null" temporaljson:"install_id,omitzero,omitempty"`
	Install        Install       `json:"-" temporaljson:"install,omitzero,omitempty"`
	Values         pgtype.Hstore `json:"values,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"values,omitzero,omitempty"`
	ValuesRedacted pgtype.Hstore `json:"redacted_values,omitzero" gorm:"type:hstore;->;-:migration" swaggertype:"object,string" temporaljson:"values_redacted,omitzero,omitempty"`

	AppInputConfigID string         `json:"app_input_config_id,omitzero" temporaljson:"app_input_config_id,omitzero,omitempty"`
	AppInputConfig   AppInputConfig `json:"-" temporaljson:"app_input_config,omitzero,omitempty"`
}

func (i *InstallInputs) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallInputs{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
