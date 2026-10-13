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

func (h *Helpers) GetInstallStateBaseline(ctx context.Context, installID string) (*state.State, error) {
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
	snapshot.Revision = generics.ToPtr(head.Revision)
	snapshot.StaleAt = nil
	snapshot.StalePartials = nil
	if !head.StaleAt.Empty() {
		snapshot.StaleAt = generics.ToPtr(head.StaleAt.Time)
		for _, partial := range head.StalePartials {
			snapshot.StalePartials = append(snapshot.StalePartials, string(partial))
		}
	}
	return snapshot, nil
}

func (h *Helpers) PublishInstallState(ctx context.Context, obj *app.InstallState, refreshed []pkgstate.PartialName) (*app.InstallState, error) {
	ctx = routing.WithoutReplica(ctx)
	if obj.State.Revision == nil {
		return nil, temporal.NewNonRetryableApplicationError("state generation needs a revision baseline", InstallStateConflict, nil)
	}
	baseline := *obj.State.Revision
	head, err := h.getInstallStateHead(ctx, h.db, obj.InstallID)
	if err != nil {
		return nil, err
	}
	if head.Revision != baseline {
		return nil, temporal.NewNonRetryableApplicationError("install state changed during generation", InstallStateConflict, nil)
	}
	if len(refreshed) == 0 {
		refreshed = pkgstate.AllPartials
	}
	remaining := make([]pkgstate.PartialName, 0, len(head.StalePartials))
	for _, partial := range head.StalePartials {
		if !slices.Contains(refreshed, partial) {
			remaining = append(remaining, partial)
		}
	}
	obj.Revision = baseline
	obj.OrgID = head.OrgID
	obj.StalePartials = remaining
	obj.StaleAt = generics.NullTime{}
	obj.State.StaleAt = nil
	obj.State.StalePartials = nil
	if len(remaining) > 0 {
		obj.StaleAt = head.StaleAt
		obj.State.StaleAt = generics.ToPtr(head.StaleAt.Time)
		for _, partial := range remaining {
			obj.State.StalePartials = append(obj.State.StalePartials, string(partial))
		}
	}
	data, err := json.Marshal(obj.State)
	if err != nil {
		return nil, err
	}
	ctx = cctx.SetOrgIDContext(ctx, head.OrgID)
	obj.StateBlob = &blobstore.Blob{}
	obj.StateBlob.Set(string(data))
	if !blobstore.IsBlobWriteEnabled(ctx) {
		return nil, fmt.Errorf("blob writes must be enabled to publish install state")
	}
	staleAt := obj.StaleAt
	if err := h.db.WithContext(ctx).Create(obj).Error; err != nil {
		return nil, err
	}
	err = h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current app.InstallStateHead
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Where(app.InstallStateHead{InstallID: obj.InstallID}).First(&current).Error; err != nil {
			return err
		}
		if current.Revision != baseline {
			return temporal.NewNonRetryableApplicationError("install state changed before publication", InstallStateConflict, nil)
		}
		publication := app.InstallStatePublication{StateID: obj.ID, InstallID: obj.InstallID, OrgID: head.OrgID, Revision: baseline + 1, CreatedAt: time.Now().UTC()}
		if err := tx.Create(&publication).Error; err != nil {
			return err
		}
		return tx.Model(&current).Select("current_state_id", "revision", "stale_at", "stale_partials").Updates(app.InstallStateHead{
			CurrentStateID: obj.ID, Revision: baseline + 1, StaleAt: staleAt, StalePartials: remaining,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	obj.State.Revision = generics.ToPtr(baseline + 1)
	return obj, nil
}
