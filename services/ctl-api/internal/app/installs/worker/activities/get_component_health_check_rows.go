package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

const (
	componentHealthCheckRowsWindow = 10 * time.Minute
	maxComponentHealthCheckRows    = 30
)

type GetComponentHealthCheckRowsRequest struct {
	InstallID          string `validate:"required"`
	InstallComponentID string `validate:"required"`
}

type ComponentHealthCheckRow struct {
	Kind         string `json:"kind" temporaljson:"kind"`
	Name         string `json:"name" temporaljson:"name"`
	Health       string `json:"health" temporaljson:"health"`
	Message      string `json:"message,omitempty" temporaljson:"message,omitempty"`
	ObservedAtTS int64  `json:"observed_at_ts,omitempty" temporaljson:"observed_at_ts,omitempty"`

	Removed bool `json:"removed,omitempty" temporaljson:"removed,omitempty"`
}

// GetComponentHealthCheckRows returns the latest observed state of every
// check/resource for one install component, for the gate's step-level narration.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) GetComponentHealthCheckRows(ctx context.Context, req *GetComponentHealthCheckRowsRequest) ([]ComponentHealthCheckRow, error) {
	var rows []app.InstallComponentResourceState
	if err := a.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(app.InstallComponentResourceStatesLatestView)).
		Select("kind", "name", "health", "message", "observed_at").
		Where(app.InstallComponentResourceState{
			InstallID:          req.InstallID,
			InstallComponentID: req.InstallComponentID,
			Source:             app.InstallComponentResourceSourceComponent,
		}).
		Where("observed_at > ?", time.Now().Add(-componentHealthCheckRowsWindow)).
		Order("kind, name").
		Limit(maxComponentHealthCheckRows).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to query component health check rows: %w", err)
	}

	out := make([]ComponentHealthCheckRow, 0, len(rows))
	for _, r := range rows {
		out = append(out, ComponentHealthCheckRow{
			Kind:         r.Kind,
			Name:         r.Name,
			Health:       r.Health,
			Message:      r.Message,
			ObservedAtTS: r.ObservedAt.Unix(),
		})
	}
	return out, nil
}
