package app

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// AppBranchRunComparison links a head branch run to its baseline (base) run and
// stores git + config diffs between them. Commit SHAs are resolved dynamically
// from each run's VCSConnectionCommit — not denormalized here.
type AppBranchRunComparison struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	HeadRunID string       `json:"head_run_id,omitzero" gorm:"not null" temporaljson:"head_run_id,omitzero,omitempty"`
	HeadRun   AppBranchRun `faker:"-" json:"head_run,omitempty" gorm:"foreignKey:HeadRunID" temporaljson:"head_run,omitzero,omitempty"`

	BaseRunID *string       `json:"base_run_id,omitempty" temporaljson:"base_run_id,omitzero,omitempty"`
	BaseRun   *AppBranchRun `faker:"-" json:"base_run,omitempty" gorm:"foreignKey:BaseRunID" temporaljson:"base_run,omitzero,omitempty"`

	GitDiff *blobstore.Blob `json:"git_diff,omitempty" temporaljson:"git_diff,omitzero,omitempty"`

	FullDiff *blobstore.Blob `json:"full_diff,omitempty" temporaljson:"full_diff,omitzero,omitempty"`

	ConfigDiff *blobstore.Blob `json:"config_diff,omitempty" temporaljson:"config_diff,omitzero,omitempty"`
}

func (c *AppBranchRunComparison) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &AppBranchRunComparison{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
		{
			Name:        indexes.Name(db, &AppBranchRunComparison{}, "head_run_id"),
			Columns:     []string{"head_run_id"},
			UniqueValue: sql.NullBool{Bool: true, Valid: true},
		},
		{
			Name: indexes.Name(db, &AppBranchRunComparison{}, "base_run_id"),
			Columns: []string{
				"base_run_id",
			},
		},
	}
}

func (c *AppBranchRunComparison) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = domains.NewAppBranchRunComparisonID()
	}

	if c.CreatedByID == "" {
		c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if c.OrgID == "" {
		c.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
