package activities

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	// componentHealthStaleAfter mirrors the evaluator's threshold: below it an
	// install is simply between reports, not quiet.
	componentHealthStaleAfter = 5 * time.Minute

	// componentHealthSweepBand bounds how far back the sweep looks. An install
	// quiet for longer than this is already unknown and re-evaluating it would
	// change nothing, so the sweep is a moving window over installs that went
	// quiet recently rather than an ever-growing tail of dead ones.
	componentHealthSweepBand = 6 * time.Hour

	// componentHealthSweepLimit caps one run so a fleet-wide outage drains over
	// several minutes instead of stampeding.
	componentHealthSweepLimit = 200

	componentHealthEvaluateSignalType = "component-health-evaluate"
)

// componentHealthSweepWindow bounds which installs a sweep considers: quiet
// long enough to count as stale, but not so long that they are already unknown.
func componentHealthSweepWindow(now time.Time) (quietBefore, ignoreBefore time.Time) {
	return now.Add(-componentHealthStaleAfter), now.Add(-componentHealthSweepBand)
}

type SweepStaleComponentHealthRequest struct{}

type SweepStaleComponentHealthResponse struct {
	Stale    int  `json:"stale"`
	Enqueued int  `json:"enqueued"`
	Capped   bool `json:"capped"`
}

// SweepStaleComponentHealth finds installs whose runner stopped reporting and
// asks each to re-evaluate, which is what turns their components unknown.
//
// This is the only part of component health that needs a clock. Verdicts for
// live installs are derived when a report arrives; silence is the one thing no
// report can tell us about, so it takes a sweep to notice.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 2m
func (a *Activities) SweepStaleComponentHealth(ctx context.Context, _ SweepStaleComponentHealthRequest) (*SweepStaleComponentHealthResponse, error) {
	resp := &SweepStaleComponentHealthResponse{}

	quietBefore, ignoreBefore := componentHealthSweepWindow(time.Now())
	var installs []app.Install
	if err := a.db.WithContext(ctx).
		Select("id").
		Where("deleted_at = 0").
		Where("last_health_report_at IS NOT NULL").
		Where("last_health_report_at < ?", quietBefore).
		Where("last_health_report_at > ?", ignoreBefore).
		Order("last_health_report_at ASC").
		Limit(componentHealthSweepLimit).
		Find(&installs).Error; err != nil {
		return nil, fmt.Errorf("unable to list stale health installs: %w", err)
	}

	resp.Stale = len(installs)
	resp.Capped = len(installs) == componentHealthSweepLimit
	if resp.Capped {
		// A silent cap reads exactly like "nothing was stale".
		a.l.Warn("component health sweep hit its per-run limit",
			zap.Int("limit", componentHealthSweepLimit))
	}

	ownerType := plugins.TableName(a.db, app.Install{})
	for _, install := range installs {
		if a.enqueueHealthEvaluate(ctx, install.ID, ownerType) {
			resp.Enqueued++
		}
	}

	return resp, nil
}

// enqueueHealthEvaluate asks one install's health queue to re-evaluate. Deduped
// per queue, so a still-pending evaluation absorbs this one.
func (a *Activities) enqueueHealthEvaluate(ctx context.Context, installID, ownerType string) bool {
	var q app.Queue
	if err := a.db.WithContext(ctx).
		Select("id").
		Where(app.Queue{
			OwnerID:   installID,
			OwnerType: ownerType,
			Name:      installshelpers.InstallComponentHealthQueueName,
		}).
		First(&q).Error; err != nil {
		return false
	}

	// Same minute-bucketing as the ingest path: a finished signal is not
	// soft-deleted until nightly cleanup, so a constant key would enqueue once
	// and then silently stop.
	dedupeKey := fmt.Sprintf("%s-%d", componentHealthEvaluateSignalType, time.Now().Truncate(time.Minute).Unix())
	if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   installID,
		OwnerType: ownerType,
		Signal: queuesignal.NewRaw(componentHealthEvaluateSignalType, map[string]any{
			"install_id": installID,
		}),
		DedupeKey: &dedupeKey,
	}); err != nil {
		a.l.Warn("unable to enqueue stale component health evaluation",
			zap.String("install_id", installID), zap.Error(err))
		return false
	}
	return true
}
