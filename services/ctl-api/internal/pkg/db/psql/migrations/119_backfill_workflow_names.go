package migrations

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const workflowNameBatchSize = 500

// these types hit neither title switch until now, so any name they already have
// is the lowercased type fallback and needs recomputing
var workflowTypesMissingTitles = []app.WorkflowType{
	app.WorkflowTypeDeprovisionSandbox,
	app.WorkflowTypeTeardownComponent,
	app.WorkflowTypeAppBranchesConfigRepoUpdate,
	app.WorkflowTypeAppBranchesComponentRepoUpdate,
	app.WorkflowTypeAppBranchConfigUpdate,
}

// Migration119BackfillWorkflowNames redoes the 108 backfill, which left name
// NULL, and recomputes the types that had no title. Unlike 108 the expression
// is not restated in SQL — names come from Workflow.ComputeName so Go stays the
// only source of truth.
func (m *Migrations) Migration119BackfillWorkflowNames(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(
		`ALTER TABLE install_workflows ADD COLUMN IF NOT EXISTS name TEXT`,
	).Error; err != nil {
		return fmt.Errorf("unable to ensure install_workflows.name: %w", err)
	}

	var batch []app.Workflow
	res := db.WithContext(ctx).
		Model(&app.Workflow{}).
		Unscoped().
		Select("id", "type", "metadata", "finished_at").
		Where("name IS NULL OR name = '' OR type IN ?", workflowTypesMissingTitles).
		FindInBatches(&batch, workflowNameBatchSize, func(_ *gorm.DB, _ int) error {
			// a fresh handle, not the batch tx — Exec appends to the
			// statement's existing Vars and the batch query's own binds
			// would still be there
			return setWorkflowNames(db.WithContext(ctx), batch)
		})
	if res.Error != nil {
		return fmt.Errorf("unable to backfill workflow names: %w", res.Error)
	}

	m.l.Info("backfilled workflow names", zap.Int64("rows", res.RowsAffected))

	if err := db.WithContext(ctx).Exec(
		`CREATE INDEX IF NOT EXISTS idx_install_workflows_name ON install_workflows (name)`,
	).Error; err != nil {
		return fmt.Errorf("unable to create idx_install_workflows_name: %w", err)
	}

	return nil
}

// one statement per batch, and never through Save/Updates — the hook would
// recompute from the partially selected struct and updated_at would move on
// every workflow in the table
func setWorkflowNames(tx *gorm.DB, batch []app.Workflow) error {
	tuples := make([]string, 0, len(batch))
	args := make([]any, 0, len(batch)*2)
	for i := range batch {
		tuples = append(tuples, "(?::text, ?::text)")
		args = append(args, batch[i].ID, batch[i].ComputeName())
	}

	sql := `UPDATE install_workflows AS w
		SET name = v.name
		FROM (VALUES ` + strings.Join(tuples, ", ") + `) AS v(id, name)
		WHERE w.id = v.id`

	if err := tx.Exec(sql, args...).Error; err != nil {
		return fmt.Errorf("unable to set names for %d workflows: %w", len(batch), err)
	}

	return nil
}
