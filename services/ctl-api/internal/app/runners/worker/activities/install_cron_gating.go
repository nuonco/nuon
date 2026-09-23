package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

const (
	installOwnerType         = "installs"
	installCronDisableReason = "no healthy runner"
	installCronEnableReason  = "runner healthy"
)

type installCronCandidate struct {
	orgID     string
	accountID string
	state     InstallCronState
}

type InstallCronState string

const (
	InstallCronsEnabled  InstallCronState = "enabled"
	InstallCronsDisabled InstallCronState = "disabled"
)

func decideInstallCronToggle(healthy bool, requested, current InstallCronState) *InstallCronState {
	target := InstallCronsDisabled
	switch {
	case healthy:
		target = InstallCronsEnabled
	case requested != InstallCronsDisabled:
		return nil
	}
	if current == target {
		return nil
	}
	return &target
}

func (a *Activities) toggleInstallCronsState(
	ctx context.Context,
	candidates map[string]*installCronCandidate,
) (*ToggleInstallCronEmitterResponse, error) {
	resp := &ToggleInstallCronEmitterResponse{}
	installIDs := make([]string, 0, len(candidates))
	for id := range candidates {
		installIDs = append(installIDs, id)
	}

	activeRunners, err := a.activeRunnerCountsByInstall(ctx, installIDs)
	if err != nil {
		return nil, err
	}

	currentStates, err := a.installCronStates(ctx, installIDs)
	if err != nil {
		return nil, err
	}

	for _, installID := range installIDs {
		candidate := candidates[installID]
		decision := decideInstallCronToggle(activeRunners[installID] > 0, candidate.state, currentStates[installID])
		if decision == nil {
			continue
		}

		fromStatus := app.StatusDisabled
		toStatus := app.StatusInProgress
		enable := *decision == InstallCronsEnabled
		reason := installCronDisableReason
		if enable {
			reason = installCronEnableReason
		} else {
			fromStatus = app.StatusInProgress
			toStatus = app.StatusDisabled
		}

		ectx := cctx.SetOrgIDContext(ctx, candidate.orgID)
		ectx = cctx.SetAccountIDContext(ectx, candidate.accountID)

		toggleResp, err := a.emitterClient.ToggleCronEmittersForOwner(ectx, &emitterclient.ToggleCronEmittersForOwnerRequest{
			OwnerID:    installID,
			OwnerType:  installOwnerType,
			FromStatus: fromStatus,
			ToStatus:   toStatus,
			Reason:     reason,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to toggle cron emitters for install %s: %w", installID, err)
		}

		if enable {
			resp.Enabled += toggleResp.Changed
		} else {
			resp.Disabled += toggleResp.Changed
		}
	}
	return resp, nil
}

type ToggleInstallCronEmitterRequest struct {
	InstallID string `validate:"required"`
	OrgID     string `validate:"required"`
	AccountID string
	State     InstallCronState `validate:"required"`
}

type ToggleInstallCronEmitterResponse struct {
	Disabled int `json:"disabled"`
	Enabled  int `json:"enabled"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) ToggleInstallCronEmitter(ctx context.Context, req ToggleInstallCronEmitterRequest) (*ToggleInstallCronEmitterResponse, error) {
	return a.toggleInstallCronsState(ctx, map[string]*installCronCandidate{
		req.InstallID: {orgID: req.OrgID, accountID: req.AccountID, state: req.State},
	})
}

func (a *Activities) activeRunnerCountsByInstall(ctx context.Context, installIDs []string) (map[string]int, error) {
	var rows []struct {
		InstallID string
		Active    int
	}
	if res := a.db.WithContext(ctx).Raw(`
		SELECT g.owner_id AS install_id, COUNT(*) FILTER (WHERE r.status = ?) AS active
		FROM runner_groups g
		JOIN runners r ON r.runner_group_id = g.id
		WHERE g.owner_type = ? AND g.owner_id IN ? AND g.deleted_at = 0 AND r.deleted_at = 0
		GROUP BY g.owner_id`,
		string(app.RunnerStatusActive),
		installOwnerType,
		installIDs,
	).Scan(&rows); res.Error != nil {
		return nil, fmt.Errorf("unable to count active runners per install: %w", res.Error)
	}

	counts := make(map[string]int, len(rows))
	for _, row := range rows {
		counts[row.InstallID] = row.Active
	}
	return counts, nil
}

func (a *Activities) installCronStates(ctx context.Context, installIDs []string) (map[string]InstallCronState, error) {
	var ownerIDs []string
	if res := a.db.WithContext(ctx).Raw(`
		SELECT DISTINCT q.owner_id
		FROM queue_emitters e
		JOIN queues q ON q.id = e.queue_id
		WHERE q.owner_type = ? AND q.owner_id IN ? AND q.deleted_at = 0
			AND e.mode = ? AND e.deleted_at = 0 AND e.status->>'status' = ?`,
		installOwnerType,
		installIDs,
		string(app.QueueEmitterModeCron),
		string(app.StatusDisabled),
	).Scan(&ownerIDs); res.Error != nil {
		return nil, fmt.Errorf("unable to list install cron states: %w", res.Error)
	}

	states := make(map[string]InstallCronState, len(installIDs))
	for _, id := range installIDs {
		states[id] = InstallCronsEnabled
	}
	for _, id := range ownerIDs {
		states[id] = InstallCronsDisabled
	}
	return states, nil
}
