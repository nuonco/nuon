package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type BuildResult struct {
	BuildID     string `json:"build_id"`
	ComponentID string `json:"component_id"`
	Status      string `json:"status"`
}

type AwaitBuildsCompleteOutput struct {
	Builds   []BuildResult `json:"builds"`
	AllDone  bool          `json:"all_done"`
	HasError bool          `json:"has_error"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
// @by-field runID
func (a *Activities) checkBuildsComplete(ctx context.Context, runID string) (*AwaitBuildsCompleteOutput, error) {
	var builds []app.ComponentBuild
	err := a.db.WithContext(ctx).
		Select("id, status, component_config_connection_id").
		Where("app_branch_run_id = ?", runID).
		Find(&builds).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get builds for run %s: %w", runID, err)
	}

	if len(builds) == 0 {
		return &AwaitBuildsCompleteOutput{AllDone: true}, nil
	}

	connIDs := make([]string, len(builds))
	for i, b := range builds {
		connIDs[i] = b.ComponentConfigConnectionID
	}

	var conns []app.ComponentConfigConnection
	a.db.WithContext(ctx).
		Select("id, component_id").
		Where("id IN ?", connIDs).
		Find(&conns)
	connMap := make(map[string]string, len(conns))
	for _, c := range conns {
		connMap[c.ID] = c.ComponentID
	}

	out := &AwaitBuildsCompleteOutput{AllDone: true}
	for _, b := range builds {
		result := BuildResult{
			BuildID:     b.ID,
			ComponentID: connMap[b.ComponentConfigConnectionID],
			Status:      string(b.Status),
		}
		out.Builds = append(out.Builds, result)

		if isTerminalBuildStatus(b.Status) {
			if b.Status != app.ComponentBuildStatusActive {
				out.HasError = true
			}
		} else {
			out.AllDone = false
		}
	}

	return out, nil
}

func isTerminalBuildStatus(s app.ComponentBuildStatus) bool {
	switch s {
	case app.ComponentBuildStatusActive, app.ComponentBuildStatusError, "cancelled":
		return true
	default:
		return false
	}
}
