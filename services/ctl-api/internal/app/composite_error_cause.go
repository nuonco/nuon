package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

// CompositeErrorCause is the edge table that wires composite errors into a
// directed acyclic causes graph.
//
// An install_deploy_failed at the top can have multiple children (one per
// failing component), each of which can in turn have its own primary cause
// (e.g. terraform_apply_failed → aws_missing_iam_permission). This shape
// supports both multi-cause aggregation and primary-cause selection in
// a single table.
//
// "At most one IsPrimary=true row per ParentID" is enforced in the helper
// layer (helpers/record.go). A partial unique DDL index is a follow-up.
//
// Delete semantics:
//
// composite_errors uses GORM soft-delete (DeletedAt) so a database-level
// ON DELETE CASCADE on these FKs would never fire — the helper read path
// (helpers.Tree) tolerates edges whose parent or child is soft-deleted by
// silently skipping rows that no longer load. Hard-delete cleanup of
// orphan edges is the responsibility of the retention sweeper (Phase 6).
type CompositeErrorCause struct {
	ParentID string          `gorm:"primaryKey;type:text;check:parent_id_checker,char_length(parent_id)=26" json:"parent_id"`
	Parent   *CompositeError `gorm:"foreignKey:ParentID;references:ID" json:"-"`

	ChildID string          `gorm:"primaryKey;type:text;check:child_id_checker,char_length(child_id)=26" json:"child_id"`
	Child   *CompositeError `gorm:"foreignKey:ChildID;references:ID" json:"-"`

	// Idx orders children under a parent for UI rendering.
	Idx int `gorm:"not null;default:0" json:"idx"`

	// IsPrimary marks the headline cause for the parent.
	IsPrimary bool `gorm:"not null;default:false" json:"is_primary"`

	CreatedAt time.Time `gorm:"notnull" json:"created_at"`
}

func (CompositeErrorCause) TableName() string { return "composite_error_causes" }

func (c *CompositeErrorCause) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &CompositeErrorCause{}, "child_id"),
			Columns: []string{"child_id"},
		},
		{
			Name:    indexes.Name(db, &CompositeErrorCause{}, "parent_id_idx"),
			Columns: []string{"parent_id", "idx"},
		},
	}
}
