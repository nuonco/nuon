package app

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type TerraformWorkspaceState struct {
	ID          string `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedBy Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	Contents []byte `json:"contents" gorm:"type:bytea" temporaljson:"contents,omitzero,omitempty"`

	Data *TerraformStateData `json:"data" temporaljson:"data,omitzero,omitempty"`

	TerraformWorkspaceID string             `temporaljson:"terraform_workspace_id,omitzero,omitempty"`
	TerraformWorkspace   TerraformWorkspace `gorm:"-" temporaljson:"terraform_workspace,omitzero,omitempty"`

	Revision int `json:"revision" gorm:"->;-:migration" temporaljson:"revision,omitzero,omitempty"`
}

func (t *TerraformWorkspaceState) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = domains.NewTerraformWorkspaceStateID()
	}

	if t.CreatedByID == "" {
		t.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if t.OrgID == "" {
		t.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (i *TerraformWorkspaceState) UseView() bool {
	return true
}

func (i *TerraformWorkspaceState) ViewVersion() string {
	return "v1"
}

func (i *TerraformWorkspaceState) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &TerraformWorkspaceState{}, 1),
			SQL:           viewsql.TerraformWorkspaceStatesViewV1,
			AlwaysReapply: true,
		},
	}
}

type TerraformStateData struct {
	Version          int                      `json:"version,omitempty" temporaljson:"version,omitzero,omitempty"`
	TerraformVersion string                   `json:"terraform_version,omitempty" temporaljson:"terraform_version,omitzero,omitempty"`
	Serial           int                      `json:"serial,omitempty" temporaljson:"serial,omitzero,omitempty"`
	Lineage          string                   `json:"lineage,omitempty" temporaljson:"lineage,omitzero,omitempty"`
	Outputs          map[string]any           `json:"outputs,omitempty" temporaljson:"outputs,omitzero,omitempty"`
	Resources        []TerraformStateResource `json:"resources,omitempty" temporaljson:"resources,omitzero,omitempty"`
	CheckResults     any                      `json:"check_results,omitempty" temporaljson:"check_results,omitzero,omitempty"`

	// base 64 encoded version of the contents for compatibility
	Contents string `json:"contents" temporaljson:"contents,omitzero,omitempty"`
}

type TerraformStateResource struct {
	Mode      string                   `json:"mode" temporaljson:"mode,omitzero,omitempty"`
	Type      string                   `json:"type" temporaljson:"type,omitzero,omitempty"`
	Name      string                   `json:"name" temporaljson:"name,omitzero,omitempty"`
	Provider  string                   `json:"provider" temporaljson:"provider,omitzero,omitempty"`
	Instances []TerraformStateInstance `json:"instances" temporaljson:"instances,omitzero,omitempty"`
}

type TerraformStateInstance struct {
	SchemaVersion       int            `json:"schema_version" temporaljson:"schema_version,omitzero,omitempty"`
	Attributes          map[string]any `json:"attributes" temporaljson:"attributes,omitzero,omitempty"`
	SensitiveAttributes []any          `json:"sensitive_attributes" temporaljson:"sensitive_attributes,omitzero,omitempty"`
}

func (c *TerraformStateData) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, c); err != nil {
			return errors.Wrap(err, "unable to scan composite status")
		}
	}
	return
}

// Value implements the driver.Valuer interface.
func (c *TerraformStateData) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (TerraformStateData) GormDataType() string {
	return "jsonb"
}
