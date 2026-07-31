package app

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
)

type TerraformWorkspaceStateJSON struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`

	// Contents is served from ContentsBlob and is no longer persisted; the legacy
	// column is dropped in a follow-up release.
	Contents     []byte          `json:"contents,omitzero" gorm:"-" temporaljson:"contents,omitzero,omitempty"`
	ContentsBlob *blobstore.Blob `json:"-" temporaljson:"-"`

	OrgID string `json:"org_id,omitzero" gorm:"default:null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

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

	if t.OrgID == "" {
		t.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if err := t.ContentsBlob.BeforeCreate(tx); err != nil {
		return err
	}

	return nil
}

// GetContents reads the state contents from the S3 blob. Returns nil when the
// blob is unset.
func (t *TerraformWorkspaceStateJSON) GetContents(ctx context.Context) ([]byte, error) {
	raw, err := t.ContentsBlob.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to read terraform workspace state json blob: %w", err)
	}
	if raw == "" {
		return nil, nil
	}

	return []byte(raw), nil
}
