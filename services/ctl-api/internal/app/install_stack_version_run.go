package app

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/viewsql"
)

type StackVersionRunType string

const (
	StackVersionRunTypeWorkflow  StackVersionRunType = "workflow-run"
	StackVersionRunTypeOutOfBand StackVersionRunType = "out-of-band-update"
)

type StackVersionRunRoleDiff struct {
	Enabled  []string `json:"enabled,omitempty"`
	Disabled []string `json:"disabled,omitempty"`
}

func (d *StackVersionRunRoleDiff) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StackVersionRunRoleDiff", value)
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d StackVersionRunRoleDiff) Value() (driver.Value, error) {
	return json.Marshal(d)
}

type StackVersionRunInputDiff struct {
	Added   []string `json:"added,omitempty"`
	Removed []string `json:"removed,omitempty"`
	Changed []string `json:"changed,omitempty"`
}

func (d *StackVersionRunInputDiff) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StackVersionRunInputDiff", value)
	}
	if len(bytes) == 0 {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

func (d StackVersionRunInputDiff) Value() (driver.Value, error) {
	return json.Marshal(d)
}

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

	RunType   StackVersionRunType       `json:"run_type,omitzero" gorm:"type:varchar(50)" temporaljson:"run_type,omitzero,omitempty"`
	RoleDiff  *StackVersionRunRoleDiff  `json:"role_diff,omitzero" gorm:"type:jsonb" temporaljson:"role_diff,omitzero,omitempty"`
	InputDiff *StackVersionRunInputDiff `json:"input_diff,omitzero" gorm:"type:jsonb" temporaljson:"input_diff,omitzero,omitempty"`
}

func (i *InstallStackVersionRun) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallStackVersionRun{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *InstallStackVersionRun) AfterQuery(tx *gorm.DB) error {
	if len(a.Data) < 1 {
		return nil
	}

	a.DataContents = map[string]any{}
	decoderConfig := &mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			generics.StringToMapDecodeHook(),
			mapstructure.StringToTimeDurationHookFunc(),
		),
		WeaklyTypedInput: true,
		Result:           &a.DataContents,
	}
	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return errors.Wrap(err, "unable to create gcp decoder")
	}
	if err := decoder.Decode(a.Data); err != nil {
		return errors.Wrap(err, "unable to parse gcp outputs")
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
