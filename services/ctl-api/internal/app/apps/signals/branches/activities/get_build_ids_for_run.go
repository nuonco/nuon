package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type BuildIDEntry struct {
	BuildID     string `json:"build_id"`
	ComponentID string `json:"component_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field runID
func (a *Activities) getBuildIDsForRun(ctx context.Context, runID string) ([]BuildIDEntry, error) {
	var builds []app.ComponentBuild
	err := a.db.WithContext(ctx).
		Select("id, component_config_connection_id").
		Where("app_branch_run_id = ?", runID).
		Find(&builds).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get builds for run %s: %w", runID, err)
	}

	result := make([]BuildIDEntry, 0, len(builds))
	for _, b := range builds {
		var conn app.ComponentConfigConnection
		if err := a.db.WithContext(ctx).
			Select("component_id").
			First(&conn, "id = ?", b.ComponentConfigConnectionID).Error; err != nil {
			continue
		}
		result = append(result, BuildIDEntry{
			BuildID:     b.ID,
			ComponentID: conn.ComponentID,
		})
	}

	return result, nil
}
