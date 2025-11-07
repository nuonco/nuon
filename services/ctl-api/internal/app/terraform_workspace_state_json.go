package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
)

type TerraformWorkspaceStateJSON struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`

	Contents []byte `json:"contents,omitzero" gorm:"type:bytea" temporaljson:"contents,omitzero,omitempty"`

	// Foreign key to TerraformWorkspace with unique constraint to prevent conflicting states for a workspace
	WorkspaceID string             `json:"workspace_id,omitzero" gorm:"type:text;not null;uniqueIndex:idx_workspace_active_lock" temporaljson:"workspace_id,omitzero,omitempty"`
	Workspace   TerraformWorkspace `json:"-" temporaljson:"workspace,omitzero,omitempty"`

	RunnerJobID *string   `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`
	RunnerJob   RunnerJob `json:"runner_job,omitzero" temporaljson:"runner_job,omitzero,omitempty"`
}

func (a *TerraformWorkspaceStateJSON) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &TerraformWorkspaceStateJSON{}, "workspace_id"),
			Columns: []string{
				"workspace_id",
			},
		},
	}
}

func (t *TerraformWorkspaceStateJSON) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = domains.NewTerraformWorkspaceStateID()
	}

	if t.CreatedByID == "" {
		t.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}
