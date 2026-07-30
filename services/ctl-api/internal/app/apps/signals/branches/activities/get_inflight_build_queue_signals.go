package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type InflightBuildQueueSignal struct {
	QueueSignalID string `json:"queue_signal_id"`
	BuildID       string `json:"build_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field runID
func (a *Activities) getInflightBuildQueueSignals(ctx context.Context, runID string) ([]InflightBuildQueueSignal, error) {
	var builds []app.ComponentBuild
	err := a.db.WithContext(ctx).
		Select("id").
		Where("app_branch_run_id = ?", runID).
		Find(&builds).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get builds for run %s: %w", runID, err)
	}

	if len(builds) == 0 {
		return nil, nil
	}

	buildIDs := make([]string, len(builds))
	for i, b := range builds {
		buildIDs[i] = b.ID
	}

	var queueSignals []app.QueueSignal
	err = a.db.WithContext(ctx).
		Select("id, owner_id").
		Where("owner_id IN ? AND owner_type = ? AND type = ? AND (status->>'status' IN (?, ?))",
			buildIDs, "component_builds", "component-build",
			string(app.StatusQueued), string(app.StatusInProgress)).
		Find(&queueSignals).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get queue signals: %w", err)
	}

	result := make([]InflightBuildQueueSignal, len(queueSignals))
	for i, qs := range queueSignals {
		result[i] = InflightBuildQueueSignal{
			QueueSignalID: qs.ID,
			BuildID:       qs.OwnerID,
		}
	}

	return result, nil
}
