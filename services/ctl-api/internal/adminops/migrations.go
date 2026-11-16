package adminops

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type MigrationsSummary struct {
	Total      int `json:"total"`
	Applied    int `json:"applied"`
	InProgress int `json:"in_progress"`
	Errored    int `json:"errored"`
}

func (o *Ops) ListMigrations(ctx context.Context) ([]migrations.MigrationModel, error) {
	var rows []migrations.MigrationModel
	if err := o.db.WithContext(ctx).
		Order("created_at ASC").
		Order("name ASC").
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to list migrations: %w", err)
	}

	return rows, nil
}

func SummarizeMigrations(rows []migrations.MigrationModel) *MigrationsSummary {
	sum := &MigrationsSummary{Total: len(rows)}
	for _, row := range rows {
		switch row.Status {
		case migrations.MigrationStatusApplied:
			sum.Applied++
		case migrations.MigrationStatusInProgress:
			sum.InProgress++
		case migrations.MigrationStatusError:
			sum.Errored++
		}
	}

	return sum
}
