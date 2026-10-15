package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"time"

	"go.temporal.io/sdk/temporal"
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

func IsInstallStateConflict(err error) bool {
	var conflict *temporal.ApplicationError
	return errors.As(err, &conflict) && conflict.Type() == InstallStateConflict
}

func (h *Helpers) getInstallStateHead(ctx context.Context, db *gorm.DB, installID string) (*app.InstallStateHead, error) {
	var head app.InstallStateHead
	err := db.WithContext(ctx).Where(app.InstallStateHead{InstallID: installID}).First(&head).Error
	if err == nil && head.CurrentStateID != "" {
		return &head, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err == gorm.ErrRecordNotFound {
			var install app.Install
			if err := tx.Select("id", "org_id").Where(app.Install{ID: installID}).First(&install).Error; err != nil {
				return err
			}
			head = app.InstallStateHead{InstallID: installID, OrgID: install.OrgID, Revision: 1}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&head).Error; err != nil {
				return err
			}
		}
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Where(app.InstallStateHead{InstallID: installID}).First(&head).Error; err != nil {
			return err
		}
		if head.CurrentStateID != "" {
			return nil
		}
		var legacy app.InstallState
		err := tx.Scopes(scopes.WithDisableViews).Where(app.InstallState{InstallID: installID}).
			Where("revision = 0").Order("created_at DESC").First(&legacy).Error
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		if err != nil {
			return err
		}
		head.CurrentStateID = legacy.ID
		if !legacy.StaleAt.Empty() {
			if head.StaleAt.Empty() {
				head.StaleAt = legacy.StaleAt
			}
			if len(legacy.StalePartials) == 0 {
				head.StalePartials = slices.Clone(pkgstate.AllPartials)
			} else {
				for _, partial := range legacy.StalePartials {
					if !slices.Contains(head.StalePartials, partial) {
						head.StalePartials = append(head.StalePartials, partial)
					}
				}
			}
		}
		return tx.Model(&head).Select("current_state_id", "stale_at", "stale_partials").Updates(head).Error
	})
	return &head, err
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
	head, err := h.getInstallStateHead(ctx, h.db, installID)
	if err != nil {
		return nil, err
	}
	snapshot := state.New()
	if head.CurrentStateID != "" {
		var row app.InstallState
		if err := h.db.WithContext(ctx).Where(app.InstallState{ID: head.CurrentStateID, InstallID: installID}).First(&row).Error; err != nil {
			return nil, err
		}
		snapshot, _ = row.GetState(ctx, h.cfg.BlobReadEnabled)
		if snapshot == nil {
			return nil, fmt.Errorf("install state %s is unavailable", row.ID)
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
	obj.Revision = baseline.Revision + 1
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
			return temporal.NewNonRetryableApplicationError("install state changed before publication", InstallStateConflict, nil)
		}
		if err := tx.Create(obj).Error; err != nil {
			return err
		}
		return tx.Model(&current).Select("current_state_id", "revision", "stale_at", "stale_partials").Updates(app.InstallStateHead{
			CurrentStateID: obj.ID, Revision: obj.Revision, StaleAt: staleAt, StalePartials: remaining,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return obj, nil
}
