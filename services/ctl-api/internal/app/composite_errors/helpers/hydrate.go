package helpers

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/pkg/composite_error/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// Hydrate turns a stored composite_errors row into the typed
// CompositeError instance via the catalog. Returns an error if the row's
// Type is not registered or the JSON does not match the type's schema.
func (h *Helpers) Hydrate(row *app.CompositeError) (composite_error.CompositeError, error) {
	if row == nil {
		return nil, errors.New("composite_errors.Hydrate: row is nil")
	}
	return catalog.Hydrate(row.Type, row.Data)
}

// Get loads a row by id and returns both the row and the typed instance.
//
// Returns (nil, nil, nil) if the row does not exist; callers that want a
// distinct "not found" should check the row pointer.
func (h *Helpers) Get(ctx context.Context, id string) (*app.CompositeError, composite_error.CompositeError, error) {
	if id == "" {
		return nil, nil, errors.New("composite_errors.Get: id is required")
	}

	var row app.CompositeError
	res := h.db.WithContext(ctx).Where("id = ?", id).First(&row)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, errors.Wrap(res.Error, "load composite_error")
	}

	typed, err := h.Hydrate(&row)
	if err != nil {
		return &row, nil, err
	}
	return &row, typed, nil
}

// ListByOwner returns every composite_error attached to the given owner,
// ordered by created_at ascending (oldest first).
//
// Severity-aware ordering is intentionally a UI concern — the dashboard
// can re-rank in the renderer if it needs to surface fatals first. Pushing
// it to SQL would lock the read path into one definition of "primary".
func (h *Helpers) ListByOwner(ctx context.Context, ownerType, ownerID string) ([]*app.CompositeError, error) {
	if ownerType == "" || ownerID == "" {
		return nil, errors.New("composite_errors.ListByOwner: owner_type and owner_id are required")
	}

	var rows []*app.CompositeError
	res := h.db.WithContext(ctx).
		Where(&app.CompositeError{OwnerType: ownerType, OwnerID: ownerID}).
		Order("created_at asc").
		Find(&rows)
	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "list composite_errors by owner")
	}

	return rows, nil
}

// Primary returns the oldest error attached to an owner — the headline error
// for status surfaces (badges, step banners). Returns (nil, nil) when none.
//
// Uses a single LIMIT 1 query instead of routing through ListByOwner so we
// do not pull every attached row into memory just to discard the rest.
func (h *Helpers) Primary(ctx context.Context, ownerType, ownerID string) (*app.CompositeError, error) {
	if ownerType == "" || ownerID == "" {
		return nil, errors.New("composite_errors.Primary: owner_type and owner_id are required")
	}

	var row app.CompositeError
	res := h.db.WithContext(ctx).
		Where(&app.CompositeError{OwnerType: ownerType, OwnerID: ownerID}).
		Order("created_at asc").
		First(&row)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(res.Error, "load primary composite_error")
	}
	return &row, nil
}
