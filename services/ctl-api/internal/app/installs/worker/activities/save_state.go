package activities

import (
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
)

type SaveStateRequest struct {
	State *state.State `validate:"required"`

	InstallID       string                         `validate:"required"`
	TriggeredByID   string                         `validate:"required"`
	TriggeredByType string                         `validate:"required"`
	GeneratedBy     app.InstallStateGenerateSource `validate:"required"`

	// RefreshedPartials, when set, carries the previous row's other stale partials forward.
	RefreshedPartials []pkgstate.PartialName
	// RefreshStartedAt keeps every marker set after the refresh began, since its data may postdate the fetch.
	RefreshStartedAt time.Time
}

// @temporal-gen-v2 activity
func (a *Activities) SaveState(ctx context.Context, req *SaveStateRequest) (result *app.InstallState, err error) {
	started := time.Now()
	defer func() { a.stateMetrics.Record(ctx, "save", started, err) }()
	// the blob upload in InstallState's BeforeCreate hook requires org_id on the context
	if keys.OrgIDFromContext(ctx) == "" {
		var install app.Install
		if res := a.db.WithContext(ctx).Select("org_id").First(&install, "id = ?", req.InstallID); res.Error != nil {
			return nil, errors.Wrap(res.Error, "unable to look up install org for state")
		}
		ctx = cctx.SetOrgIDContext(ctx, install.OrgID)
	}

	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := helpers.LockInstallInputs(ctx, tx, req.InstallID); err != nil {
			return err
		}
		inputs, err := a.helpers.GetInstallInputsState(ctx, tx, req.InstallID, false, false)
		if err != nil {
			return err
		}
		snapshot := *req.State
		snapshot.Inputs = inputs
		helpers.MapLegacyFields(&snapshot)
		stateJSON, err := json.Marshal(&snapshot)
		if err != nil {
			return errors.Wrap(err, "unable to marshal install state for blob")
		}

		obj := &app.InstallState{
			InstallID:       req.InstallID,
			TriggeredByID:   req.TriggeredByID,
			TriggeredByType: req.TriggeredByType,
			State:           &snapshot,
			GeneratedBy:     req.GeneratedBy,
			StateBlob:       &blobstore.Blob{},
		}
		obj.StateBlob.Set(string(stateJSON))
		if len(req.RefreshedPartials) > 0 {
			carried, err := a.carriedStalePartials(ctx, tx, req.InstallID, req.RefreshedPartials, req.RefreshStartedAt)
			if err != nil {
				return err
			}
			if len(carried) > 0 {
				obj.StaleAt = generics.NewNullTime(time.Now())
				obj.StalePartials = carried
			}
		}
		if err := tx.Create(obj).Error; err != nil {
			return errors.Wrap(err, "unable to create install state")
		}
		result = obj
		return nil
	})
	return result, err
}

func (a *Activities) carriedStalePartials(ctx context.Context, db *gorm.DB, installID string, refreshed []pkgstate.PartialName, refreshStartedAt time.Time) ([]pkgstate.PartialName, error) {
	var prev app.InstallState
	res := db.WithContext(ctx).
		Select("id", "stale_at", "stale_partials").
		Where(app.InstallState{InstallID: installID}).
		Order("created_at DESC").
		Limit(1).
		Find(&prev)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get previous install state")
	}

	if !prev.StaleAt.Empty() && !refreshStartedAt.IsZero() && prev.StaleAt.Time.After(refreshStartedAt) {
		return prev.StalePartials, nil
	}

	var carried []pkgstate.PartialName
	for _, p := range prev.StalePartials {
		if !slices.Contains(refreshed, p) {
			carried = append(carried, p)
		}
	}
	return carried, nil
}
