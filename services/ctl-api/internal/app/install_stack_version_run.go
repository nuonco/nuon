package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/pkg/types/stacks"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type InstallStackVersionRun struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallStackVersionID string              `json:"install_stack_version_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"install_stack_version_id,omitzero,omitempty"`
	InstallStackVersion   InstallStackVersion `json:"-" temporaljson:"install_stack_version,omitzero,omitempty"`

	Data         pgtype.Hstore  `json:"data,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"data,omitzero,omitempty"`
	DataContents map[string]any `json:"data_contents,omitzero" gorm:"-"`
}

func (a *InstallStackVersionRun) AfterQuery(tx *gorm.DB) error {
	if len(a.Data) < 1 {
		return nil
	}

	// TODO(ja): what have i become
	_, isAzure := a.Data["resource_group_id"]
	if !isAzure {
		// parsing pgtype.Hstore into map[string]interface{}
		outputData, err := stacks.DecodeAWSStackOutputData(a.Data)
		if err != nil {
			return errors.Wrap(err, "unable to decode stack output data to map")
		}
		a.DataContents = outputData

	}

	return nil
}

func (i *InstallStackVersionRun) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallStackVersionRunID()
	}

	if i.CreatedByID == "" {
		i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if i.OrgID == "" {
		i.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (i *InstallStackVersionRun) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.CustomViewName(db, &InstallStackVersionRun{}, "state_view_v1"),
			SQL:           viewsql.InstallStackVersionRunsStateViewV1,
			AlwaysReapply: true,
		},
	}
}
