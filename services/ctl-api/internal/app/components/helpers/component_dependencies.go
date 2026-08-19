package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (h *Helpers) ClearComponentDependencies(ctx context.Context, compID string) error {
	return h.ClearComponentDependenciesWithDB(ctx, h.db, compID)
}

func (h *Helpers) ClearComponentDependenciesWithDB(ctx context.Context, db *gorm.DB, compID string) error {
	// clear dependencies
	compDep := app.ComponentDependency{}
	res := db.WithContext(ctx).
		Unscoped().
		Delete(&compDep, "component_id = ?", compID)
	if res.Error != nil {
		return fmt.Errorf("unable to clear component dependencies: %w", res.Error)
	}

	return nil
}

// NOTE: GORM does not support callbacks when using a custom join table on many2many relationships + associations mode,
// so this is a helper used to create component dependencies
func (h *Helpers) CreateComponentDependencies(ctx context.Context, compID string, dependencyIDs []string) error {
	return h.CreateComponentDependenciesWithDB(ctx, h.db, compID, dependencyIDs)
}

// Callers inside a transaction must use this, or the component FKs fail.
func (h *Helpers) CreateComponentDependenciesWithDB(ctx context.Context, db *gorm.DB, compID string, dependencyIDs []string) error {
	if len(dependencyIDs) < 1 {
		return nil
	}

	// create dependencies
	deps := make([]*app.ComponentDependency, 0, len(dependencyIDs))
	for _, depID := range dependencyIDs {
		deps = append(deps, &app.ComponentDependency{
			ComponentID:  compID,
			DependencyID: depID,
		})
	}

	res := db.WithContext(ctx).
		Create(&deps)
	if res.Error != nil {
		return res.Error
	}

	return nil
}
