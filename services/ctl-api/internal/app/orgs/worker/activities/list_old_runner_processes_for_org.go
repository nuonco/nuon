package activities

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ListOldRunnerProcessesForOrgRequest struct {
	OrgID string `validate:"required"`
}

type StaleRunnerProcess struct {
	ID       string `json:"id"`
	RunnerID string `json:"runner_id"`
}

type ListOldRunnerProcessesForOrgResponse struct {
	Processes []StaleRunnerProcess `json:"processes"`
}

// @temporal-gen-v2 activity
// @by-field OrgID
//
// ListOldRunnerProcessesForOrg returns every runner process for the org except
// the most recent per (runner_id, type). Used by the
// org-remove-old-runner-processes signal to drive per-process termination.
func (a *Activities) ListOldRunnerProcessesForOrg(ctx context.Context, req ListOldRunnerProcessesForOrgRequest) (*ListOldRunnerProcessesForOrgResponse, error) {
	var processes []app.RunnerProcess
	if res := a.db.WithContext(ctx).
		Where("org_id = ?", req.OrgID).
		Order("runner_id, type, created_at DESC").
		Find(&processes); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to list runner processes")
	}

	type key struct {
		RunnerID string
		Type     app.RunnerProcessType
	}
	seen := make(map[key]bool)
	stale := make([]StaleRunnerProcess, 0)
	for _, p := range processes {
		k := key{RunnerID: p.RunnerID, Type: p.Type}
		if seen[k] {
			stale = append(stale, StaleRunnerProcess{ID: p.ID, RunnerID: p.RunnerID})
		} else {
			seen[k] = true
		}
	}

	return &ListOldRunnerProcessesForOrgResponse{Processes: stale}, nil
}
