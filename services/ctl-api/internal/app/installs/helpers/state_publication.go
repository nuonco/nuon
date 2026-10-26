package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

const InstallStateConflict = "InstallStateConflict"

const (
	installStateArchiveRetention = 50
	installStateArchiveBatch     = 100
)

func IsInstallStateConflict(err error) bool {
	var conflict *temporal.ApplicationError
	return errors.As(err, &conflict) && conflict.Type() == InstallStateConflict
}

type InstallStateBaseline struct {
	State *state.State
	Head  *app.InstallStateHead
}

func (h *Helpers) waitForInstallStateSources(ctx context.Context, orgID string) (err error) {
	started := time.Now()
	waitCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	defer func() {
		h.stateMetrics.Record(ctx, "source_wait", started, err)
		if err != nil && errors.Is(waitCtx.Err(), context.DeadlineExceeded) && ctx.Err() == nil {
			h.stateMetrics.RecordSourceWaitTimeout(ctx)
		}
	}()
	for {
		var ready bool
		if err := h.db.WithContext(waitCtx).Raw(
			"SELECT pg_try_advisory_xact_lock(hashtext('install_state_sources'), hashtext(?))", orgID,
		).Scan(&ready).Error; err != nil {
			return err
		}
		if ready {
			return nil
		}
		select {
		case <-waitCtx.Done():
			return waitCtx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (h *Helpers) GetInstallStateBaseline(ctx context.Context, installID string) (*InstallStateBaseline, error) {
	ctx = routing.WithoutReplica(ctx)
	var row struct {
		app.InstallStateHead
		Snapshot app.InstallState `gorm:"embedded;embeddedPrefix:snapshot_"`
	}
	if err := h.db.WithContext(ctx).Model(&app.InstallStateHead{}).
		Select("install_state_heads.*, snapshot.id AS snapshot_id, snapshot.state AS snapshot_state, snapshot.state_blob AS snapshot_state_blob").
		Joins("LEFT JOIN install_states snapshot ON snapshot.id = install_state_heads.current_state_id AND snapshot.install_id = install_state_heads.install_id AND snapshot.deleted_at = 0").
		Where(app.InstallStateHead{InstallID: installID}).Take(&row).Error; err != nil {
		return nil, err
	}
	head := &row.InstallStateHead
	snapshot := state.New()
	if head.CurrentStateID != "" {
		snapshot, _ = row.Snapshot.GetState(ctx, h.cfg.BlobReadEnabled)
		if snapshot == nil {
			return nil, fmt.Errorf("install state %s is unavailable", head.CurrentStateID)
		}
	}
	snapshot.StaleAt = nil
	if !head.StaleAt.Empty() {
		snapshot.StaleAt = generics.ToPtr(head.StaleAt.Time)
	}
	return &InstallStateBaseline{State: snapshot, Head: head}, nil
}

func (h *Helpers) PublishInstallState(ctx context.Context, obj *app.InstallState, baseline *app.InstallStateHead, refreshed []pkgstate.PartialName) (*app.InstallState, error) {
	ctx = routing.WithoutReplica(ctx)
	remaining := make([]pkgstate.PartialName, 0, len(baseline.StalePartials))
	for _, partial := range baseline.StalePartials {
		if !slices.Contains(refreshed, partial) {
			remaining = append(remaining, partial)
		}
	}
	staleAt := generics.NullTime{}
	if len(remaining) > 0 {
		staleAt = baseline.StaleAt
	}
	obj.OrgID = baseline.OrgID
	obj.State.StaleAt = nil
	data, err := json.Marshal(obj.State)
	if err != nil {
		return nil, err
	}
	ctx = cctx.SetOrgIDContext(ctx, baseline.OrgID)
	obj.StateBlob = &blobstore.Blob{}
	obj.StateBlob.Set(string(data))
	if !blobstore.IsBlobWriteEnabled(ctx) {
		return nil, fmt.Errorf("blob writes must be enabled to publish install state")
	}
	if err := obj.StateBlob.BeforeCreate(h.db.WithContext(ctx)); err != nil {
		return nil, err
	}
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var install app.Install
		if err := tx.Scopes(scopes.WithDisableViews).Clauses(clause.Locking{Strength: "KEY SHARE"}).Select("id").Where(app.Install{ID: obj.InstallID}).First(&install).Error; err != nil {
			return err
		}
		var current app.InstallStateHead
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Where(app.InstallStateHead{InstallID: obj.InstallID}).First(&current).Error; err != nil {
			return err
		}
		if current.Revision != baseline.Revision {
			return temporal.NewApplicationError("install state changed before publication", InstallStateConflict)
		}
		if err := tx.Create(obj).Error; err != nil {
			return err
		}
		return tx.Model(&current).Select("current_state_id", "revision", "stale_at", "stale_partials").Updates(app.InstallStateHead{
			CurrentStateID: obj.ID, Revision: baseline.Revision + 1, StaleAt: staleAt, StalePartials: remaining,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	if err := h.ArchiveInstallState(ctx, obj.InstallID); err != nil {
		cctx.GetLogger(ctx, h.l).Warn("unable to archive install state", zap.String("install_id", obj.InstallID), zap.Error(err))
	}
	return obj, nil
}

func (h *Helpers) ArchiveInstallState(ctx context.Context, installID string) error {
	// The keep-list is a finished set, then a partial-index lookup locks a bounded
	// batch. NOT IN (ORDER BY ... LIMIT) on the update made Postgres scan the table
	// and block the publish for about 30s.
	return h.db.WithContext(routing.WithoutReplica(ctx)).Exec(`
UPDATE install_states
SET archived = TRUE, state = NULL, updated_at = ?
WHERE id IN (
	SELECT candidate.id
	FROM install_states AS candidate
	WHERE candidate.install_id = ?
		AND candidate.deleted_at = 0
		AND NOT candidate.archived
		AND candidate.id NOT IN (
			SELECT keep.id
			FROM (
				SELECT id
				FROM install_states
				WHERE install_id = ? AND deleted_at = 0
				ORDER BY created_at DESC, id DESC
				LIMIT ?
			) AS keep
		)
		AND candidate.id IS DISTINCT FROM (
			SELECT current_state_id
			FROM install_state_heads
			WHERE install_id = ?
		)
	LIMIT ?
	FOR UPDATE SKIP LOCKED
)`, time.Now().UTC(), installID, installID, installStateArchiveRetention, installID, installStateArchiveBatch).Error
}
