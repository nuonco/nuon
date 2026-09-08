package activities

import (
	"context"
	"fmt"
	"sort"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
)

const (
	installOwnerType         = "installs"
	installCronDisableReason = "no healthy runner"
	installCronEnableReason  = "runner healthy"
)

// installCronCandidate tracks what one page of runners implies for an install.
type installCronCandidate struct {
	orgID     string
	accountID string
	disable   bool
}

type cronGateAction int

const (
	cronGateNoop cronGateAction = iota
	cronGateDisable
	cronGateEnable
)

// decideCronGate resolves one install's cron state. It is edge-triggered on
// what is actually persisted rather than on the health transition, so a
// converged install costs nothing and an install whose offline runner was
// replaced rather than recovered still gets its crons back.
func decideCronGate(healthy, disableCandidate, currentlyDisabled bool) cronGateAction {
	switch {
	case healthy && currentlyDisabled:
		return cronGateEnable
	case !healthy && disableCandidate && !currentlyDisabled:
		return cronGateDisable
	default:
		return cronGateNoop
	}
}

// applyInstallCronGating switches an install's cron emitters off once none of
// its runners are healthy, and back on as soon as one is.
//
// The healthy-runner count is resolved against the whole runner group rather
// than the current page, so a group split across pages cannot flap. Re-enabling
// keys off which emitters are actually disabled instead of a recovery
// transition, so an install whose offline runner was replaced rather than
// recovered still gets its crons back.
func (a *Activities) applyInstallCronGating(
	ctx context.Context,
	candidates map[string]*installCronCandidate,
	resp *BatchRunnerHealthchecksResponse,
) {
	installIDs := make([]string, 0, len(candidates))
	for id := range candidates {
		installIDs = append(installIDs, id)
	}
	sort.Strings(installIDs)

	activeRunners, err := a.activeRunnerCountsByInstall(ctx, installIDs)
	if err != nil {
		resp.Errors += len(installIDs)
		a.l.Warn("unable to count active runners for install cron gating", zap.Error(err))
		return
	}

	disabled, err := a.installsWithDisabledCrons(ctx, installIDs)
	if err != nil {
		resp.Errors += len(installIDs)
		a.l.Warn("unable to resolve disabled install crons", zap.Error(err))
		return
	}

	for _, installID := range installIDs {
		c := candidates[installID]

		action := decideCronGate(activeRunners[installID] > 0, c.disable, disabled[installID])
		if action == cronGateNoop {
			continue
		}

		enable := action == cronGateEnable
		reason := installCronDisableReason
		if enable {
			reason = installCronEnableReason
		}

		ectx := cctx.SetOrgIDContext(ctx, c.orgID)
		ectx = cctx.SetAccountIDContext(ectx, c.accountID)

		setResp, err := a.emitterClient.SetCronEmittersEnabledForOwner(ectx, &emitterclient.SetCronEmittersEnabledRequest{
			OwnerID:   installID,
			OwnerType: installOwnerType,
			Enabled:   enable,
			Reason:    reason,
		})
		if err != nil {
			resp.Errors++
			a.l.Warn("unable to gate install cron emitters",
				zap.String("install_id", installID),
				zap.Bool("enabled", enable),
				zap.Error(err))
			continue
		}

		resp.Errors += setResp.Errors
		if enable {
			resp.CronsEnabled += setResp.Changed
		} else {
			resp.CronsDisabled += setResp.Changed
		}
	}
}

type GateInstallCronEmittersRequest struct {
	InstallID string `validate:"required"`
	OrgID     string `validate:"required"`
	AccountID string

	// Disable asks for the install's crons to be switched off. It is ignored
	// when any runner in the install's group is still healthy.
	Disable bool
}

type GateInstallCronEmittersResponse struct {
	Disabled int `json:"disabled"`
	Enabled  int `json:"enabled"`
	Errors   int `json:"errors"`
}

// GateInstallCronEmitters applies the same rule as applyInstallCronGating to a
// single install, for the per-runner healthcheck signal that orgs without
// org-healthcheck-sweeps still run.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) GateInstallCronEmitters(ctx context.Context, req GateInstallCronEmittersRequest) (*GateInstallCronEmittersResponse, error) {
	resp := &GateInstallCronEmittersResponse{}
	batch := &BatchRunnerHealthchecksResponse{}

	a.applyInstallCronGating(ctx, map[string]*installCronCandidate{
		req.InstallID: {orgID: req.OrgID, accountID: req.AccountID, disable: req.Disable},
	}, batch)

	resp.Disabled = batch.CronsDisabled
	resp.Enabled = batch.CronsEnabled
	resp.Errors = batch.Errors
	return resp, nil
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

func (a *Activities) installsWithDisabledCrons(ctx context.Context, installIDs []string) (map[string]bool, error) {
	var ownerIDs []string
	if res := a.db.WithContext(ctx).Raw(`
		SELECT DISTINCT q.owner_id
		FROM queue_emitters e
		JOIN queues q ON q.id = e.queue_id
		WHERE q.owner_type = ? AND q.owner_id IN ? AND q.deleted_at = 0
			AND e.mode = ? AND e.deleted_at = 0 AND e.enabled = false`,
		installOwnerType,
		installIDs,
		string(app.QueueEmitterModeCron),
	).Scan(&ownerIDs); res.Error != nil {
		return nil, fmt.Errorf("unable to list installs with disabled crons: %w", res.Error)
	}

	disabled := make(map[string]bool, len(ownerIDs))
	for _, id := range ownerIDs {
		disabled[id] = true
	}
	return disabled, nil
}
