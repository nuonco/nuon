package adminops

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/restart"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

const (
	enqueueSourceRunnerRestart = "admin-runners-update"
	runnerOwnerType            = "runners"
	runnerGroupPageSize        = 100
	runnerUpdateConcurrency    = 10
)

type RunnerUpdateFailure struct {
	RunnerID string `json:"runner_id"`
	Error    string `json:"error"`
}

type UpdateRunnersOptions struct {
	DryRun        bool
	RestartSignal bool
}

type UpdateRunnersResult struct {
	Tag            string                `json:"tag"`
	DryRun         bool                  `json:"dry_run"`
	RestartSignal  bool                  `json:"restart_signal"`
	Total          int                   `json:"total"`
	Updated        int                   `json:"updated"`
	Skipped        int                   `json:"skipped"`
	Failed         int                   `json:"failed"`
	RestartSkipped int                   `json:"restart_skipped,omitempty"`
	RunnerIDs      []string              `json:"runner_ids,omitempty"`
	Failures       []RunnerUpdateFailure `json:"failures,omitempty"`

	mu sync.Mutex
}

func (o *Ops) UpdateInstallRunners(ctx context.Context, tag string, opts UpdateRunnersOptions) (*UpdateRunnersResult, error) {
	res := &UpdateRunnersResult{Tag: tag, DryRun: opts.DryRun, RestartSignal: opts.RestartSignal}

	after := ""
	for {
		groups, err := o.installRunnerGroups(ctx, after)
		if err != nil {
			return res, err
		}
		if len(groups) == 0 {
			break
		}
		after = groups[len(groups)-1].ID

		var eg errgroup.Group
		eg.SetLimit(runnerUpdateConcurrency)
		for i := range groups {
			group := &groups[i]
			eg.Go(func() error {
				o.updateRunnerGroup(ctx, group, tag, opts, res)
				return nil
			})
		}
		_ = eg.Wait()

		if err := ctx.Err(); err != nil {
			return res, fmt.Errorf("stopped updating install runners: %w", err)
		}
	}

	if res.Failed > 0 {
		return res, fmt.Errorf("unable to update %d of %d install runners", res.Failed, res.Total)
	}

	return res, nil
}

func (o *Ops) installRunnerGroups(ctx context.Context, after string) ([]app.RunnerGroup, error) {
	var groups []app.RunnerGroup
	res := o.db.WithContext(ctx).
		Preload("Runners").
		Preload("Settings").
		Where(app.RunnerGroup{Type: app.RunnerGroupTypeInstall}).
		Where(clause.Gt{Column: clause.Column{Name: "id"}, Value: after}).
		Order("id ASC").
		Limit(runnerGroupPageSize).
		Find(&groups)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to list install runner groups: %w", res.Error)
	}

	return groups, nil
}

func (o *Ops) updateRunnerGroup(ctx context.Context, group *app.RunnerGroup, tag string, opts UpdateRunnersOptions, res *UpdateRunnersResult) {
	if len(group.Runners) == 0 {
		return
	}

	if group.Settings.ContainerImageTag == tag && group.Settings.BinaryVersion == tag {
		res.record(func() {
			res.Total += len(group.Runners)
			res.Skipped += len(group.Runners)
		})
		return
	}

	if opts.DryRun {
		res.record(func() {
			res.Total += len(group.Runners)
			res.Updated += len(group.Runners)
			for _, runner := range group.Runners {
				res.RunnerIDs = append(res.RunnerIDs, runner.ID)
			}
		})
		return
	}

	ctx = cctx.SetOrgIDContext(ctx, group.OrgID)

	if err := o.updateRunnerGroupSettings(ctx, group.ID, tag); err != nil {
		res.record(func() {
			res.Total += len(group.Runners)
			for _, runner := range group.Runners {
				res.fail(runner.ID, err)
			}
		})
		return
	}

	if !opts.RestartSignal {
		res.record(func() {
			res.Total += len(group.Runners)
			res.Updated += len(group.Runners)
		})
		return
	}

	for _, runner := range group.Runners {
		sent, err := o.restartRunner(ctx, runner.ID)
		res.record(func() {
			res.Total++
			if err != nil {
				res.fail(runner.ID, fmt.Errorf("settings updated: %w", err))
				return
			}
			res.Updated++
			if !sent {
				res.RestartSkipped++
			}
		})
	}
}

func (o *Ops) updateRunnerGroupSettings(ctx context.Context, runnerGroupID, tag string) error {
	res := o.db.WithContext(ctx).
		Model(&app.RunnerGroupSettings{}).
		Where(app.RunnerGroupSettings{RunnerGroupID: runnerGroupID}).
		Updates(app.RunnerGroupSettings{ContainerImageTag: tag, BinaryVersion: tag})
	if res.Error != nil {
		return fmt.Errorf("unable to update runner group settings: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("runner group %s has no settings", runnerGroupID)
	}

	return nil
}

func (o *Ops) restartRunner(ctx context.Context, runnerID string) (bool, error) {
	var q app.Queue
	res := o.db.WithContext(ctx).
		Where(app.Queue{
			OwnerID:   runnerID,
			OwnerType: runnerOwnerType,
			Name:      queuenames.RunnerSignalsQueueName,
		}).
		First(&q)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		zap.L().Warn("runner has no signals queue, skipping restart signal", zap.String("runner_id", runnerID))
		return false, nil
	}
	if res.Error != nil {
		return false, fmt.Errorf("unable to get runner signals queue: %w", res.Error)
	}

	resp, err := o.queueClient.EnqueueSignalInTransaction(ctx, o.db, &queueclient.EnqueueSignalRequest{
		QueueID:   q.ID,
		OwnerID:   runnerID,
		OwnerType: runnerOwnerType,
		Signal:    &restart.Signal{RunnerID: runnerID},
	})
	if err != nil {
		return false, fmt.Errorf("unable to enqueue restart signal: %w", err)
	}

	if err := o.enqueuer.EnqueueInline(ctx, resp.ID, enqueueSourceRunnerRestart); err != nil {
		return false, fmt.Errorf("unable to dispatch restart signal: %w", err)
	}

	return true, nil
}

func (r *UpdateRunnersResult) record(update func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	update()
}

func (r *UpdateRunnersResult) fail(runnerID string, err error) {
	r.Failed++
	r.Failures = append(r.Failures, RunnerUpdateFailure{RunnerID: runnerID, Error: err.Error()})
}
