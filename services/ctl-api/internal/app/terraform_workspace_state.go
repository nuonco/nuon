package app

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
)

type TerraformWorkspaceState struct {
	ID          string `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`

	CreatedBy Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	Contents []byte `json:"contents,omitzero" gorm:"type:bytea" temporaljson:"contents,omitzero,omitempty"`

	TerraformWorkspaceID string             `json:"terraform_workspace_id,omitzero" temporaljson:"terraform_workspace_id,omitzero,omitempty"`
	TerraformWorkspace   TerraformWorkspace `json:"terraform_workspace,omitzero" gorm:"-" temporaljson:"terraform_workspace,omitzero,omitempty"`

	RunnerJobID *string   `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerJob   RunnerJob `json:"runner_job,omitzero" temporaljson:"runner_job,omitzero,omitempty"`

	Revision int `json:"revision,omitzero" gorm:"->;-:migration" temporaljson:"revision,omitzero,omitempty"`
}

func (t *TerraformWorkspaceState) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &TerraformWorkspaceState{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
	Version          int                      `json:"version,omitzero,omitempty" temporaljson:"version,omitzero,omitempty"`
	TerraformVersion string                   `json:"terraform_version,omitzero,omitempty" temporaljson:"terraform_version,omitzero,omitempty"`
	Serial           int                      `json:"serial,omitzero,omitempty" temporaljson:"serial,omitzero,omitempty"`
	Lineage          string                   `json:"lineage,omitzero,omitempty" temporaljson:"lineage,omitzero,omitempty"`
	Outputs          map[string]any           `json:"outputs,omitzero,omitempty" temporaljson:"outputs,omitzero,omitempty"`
	Resources        []TerraformStateResource `json:"resources,omitzero,omitempty" temporaljson:"resources,omitzero,omitempty"`
	CheckResults     any                      `json:"check_results,omitzero,omitempty" temporaljson:"check_results,omitzero,omitempty"`

	// base 64 encoded version of the contents for compatibility
	Contents string `json:"contents,omitzero" temporaljson:"contents,omitzero,omitempty"`
}

type TerraformStateResource struct {
	Mode      string                   `json:"mode,omitzero" temporaljson:"mode,omitzero,omitempty"`
	Type      string                   `json:"type,omitzero" temporaljson:"type,omitzero,omitempty"`
	Name      string                   `json:"name,omitzero" temporaljson:"name,omitzero,omitempty"`
	Provider  string                   `json:"provider,omitzero" temporaljson:"provider,omitzero,omitempty"`
	Instances []TerraformStateInstance `json:"instances,omitzero" temporaljson:"instances,omitzero,omitempty"`
}

type TerraformStateInstance struct {
	SchemaVersion       int            `json:"schema_version,omitzero" temporaljson:"schema_version,omitzero,omitempty"`
	Attributes          map[string]any `json:"attributes,omitzero" temporaljson:"attributes,omitzero,omitempty"`
	SensitiveAttributes []any          `json:"sensitive_attributes,omitzero" temporaljson:"sensitive_attributes,omitzero,omitempty"`
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
