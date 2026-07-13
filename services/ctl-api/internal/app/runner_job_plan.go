package app

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerJobPlan struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_job_plan,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerJobID string `json:"runner_job_id,omitzero" gorm:"defaultnull;notnull;index:idx_runner_job_plan,unique" temporaljson:"runner_job_id,omitzero,omitempty"`

	PlanJSON string `json:"plan_json,omitzero" temporaljson:"plan_json,omitzero,omitempty"`
	// Deprecated: composite plans are read from CompositePlanBlob (S3). This
	// jsonb column is retained only as a fallback for rows not yet backfilled.
	CompositePlan     plantypes.CompositePlan `json:"composite_plan,omitzero" gorm:"type:jsonb" temporaljson:"composite_plan,omitzero,omitempty"`
	CompositePlanBlob *blobstore.Blob         `json:"-" temporaljson:"-"`
}

func (r *RunnerJobPlan) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &RunnerJobPlan{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (r *RunnerJobPlan) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerID()
	}

	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	if err := r.CompositePlanBlob.BeforeCreate(tx); err != nil {
		return err
	}

	return nil
}

// GetCompositePlan returns the composite plan. When blobRead is enabled it reads
// from the S3 blob, falling back to the legacy jsonb column when the blob is
// unset or unreadable (rows not yet backfilled). When disabled it always reads
// the legacy column. The second return reports whether the plan came from the
// blob.
func (r *RunnerJobPlan) GetCompositePlan(ctx context.Context, blobRead bool) (*plantypes.CompositePlan, bool) {
	if blobRead {
		if raw, err := r.CompositePlanBlob.Get(ctx); err == nil && raw != "" {
			var cp plantypes.CompositePlan
			if err := json.Unmarshal([]byte(raw), &cp); err == nil {
				return &cp, true
			}
		}
	}

	return &r.CompositePlan, false
}

func (r *RunnerJobPlan) DeriveCompositePlan(runnerJob *RunnerJob) (*plantypes.CompositePlan, error) {
	var compositePlan plantypes.CompositePlan
	switch runnerJob.Group {
	case RunnerJobGroupSync:
		switch runnerJob.Type {
		case RunnerJobTypeOCISync, RunnerJobTypeNOOPSync:
			if err := json.Unmarshal([]byte(r.PlanJSON), &compositePlan.SyncOCIPlan); err != nil {
				return nil, fmt.Errorf("unable to unmarshal sync oci plan: %w", err)
			}
		case RunnerJobTypeSandboxSyncSecrets:
			if err := json.Unmarshal([]byte(r.PlanJSON), &compositePlan.SyncSecretsPlan); err != nil {
				return nil, fmt.Errorf("unable to unmarshal sync secret plan: %w", err)
			}
		case RunnerJobTypeFetchImageMetadata:
			if err := json.Unmarshal([]byte(r.PlanJSON), &compositePlan.FetchImageMetadataPlan); err != nil {
				return nil, fmt.Errorf("unable to unmarshal fetch image metadata plan: %w", err)
			}
		default:
			return nil, fmt.Errorf("unknown sync job type: %s", runnerJob.Type)
		}
	case RunnerJobGroupBuild:
		if err := json.Unmarshal([]byte(r.PlanJSON), &compositePlan.BuildPlan); err != nil {
			return nil, fmt.Errorf("unable to unmarshal build plan: %w", err)
		}
	case RunnerJobGroupDeploy:
		if err := json.Unmarshal([]byte(r.PlanJSON), &compositePlan.DeployPlan); err != nil {
			return nil, fmt.Errorf("unable to unmarshal deploy plan: %w", err)
		}
	case RunnerJobGroupActions:
		if err := json.Unmarshal([]byte(r.PlanJSON), &compositePlan.ActionWorkflowRunPlan); err != nil {
			return nil, fmt.Errorf("unable to unmarshal action plan: %w", err)
		}
	case RunnerJobGroupSandbox:
		if err := json.Unmarshal([]byte(r.PlanJSON), &compositePlan.SandboxRunPlan); err != nil {
			return nil, fmt.Errorf("unable to unmarshal sandbox plan: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown runner job group: %s", runnerJob.Group)
	}

	return &compositePlan, nil
}
