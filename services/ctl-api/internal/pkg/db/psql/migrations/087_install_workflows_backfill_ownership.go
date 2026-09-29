package migrations

import (
	"context"
	_ "embed"

	"gorm.io/gorm"
)

func (m *Migrations) Migration087InstallWorkflowsBackfillOwnership(ctx context.Context, db *gorm.DB) error {
	return nil
}
