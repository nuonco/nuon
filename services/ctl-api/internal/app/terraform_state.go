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

type TerraformState struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id"`
	Org   Org    `json:"-"`

	Data *TerraformStateData `json:"data"`
	Lock *TerraformLock      `json:"lock"`

	TerraformWorkspaceID string
	TerraformWorkspace   TerraformWorkspace `gorm:"-"`

	Revision int `json:"revision" gorm:"->;-:migration"`
}

func (t *TerraformState) BeforeCreate(tx *gorm.DB) (err error) {
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

func (i *TerraformState) UseView() bool {
	return true
}

func (i *TerraformState) ViewVersion() string {
	return "v1"
}

func (i *TerraformState) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &TerraformState{}, 1),
			SQL:           viewsql.TerraformStatesViewV1,
			AlwaysReapply: true,
		},
	}
}

type TerraformStateData struct {
	Version          int                      `json:"version,omitempty"`
	TerraformVersion string                   `json:"terraform_version,omitempty"`
	Serial           int                      `json:"serial,omitempty"`
	Lineage          string                   `json:"lineage,omitempty"`
	Outputs          map[string]any           `json:"outputs,omitempty"`
	Resources        []TerraformStateResource `json:"resources,omitempty"`
	CheckResults     any                      `json:"check_results,omitempty"`
}

type TerraformStateResource struct {
	Mode      string                   `json:"mode"`
	Type      string                   `json:"type"`
	Name      string                   `json:"name"`
	Provider  string                   `json:"provider"`
	Instances []TerraformStateInstance `json:"instances"`
}

type TerraformStateInstance struct {
	SchemaVersion       int            `json:"schema_version"`
	Attributes          map[string]any `json:"attributes"`
	SensitiveAttributes []any          `json:"sensitive_attributes"`
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
