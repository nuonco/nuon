package helpers

import (
	"context"
	"fmt"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// StackPhoneHomeRequest is the free-form payload an install stack reports: the
// stack's outputs plus a `request_type` naming the lifecycle event.
type StackPhoneHomeRequest map[string]any

// The lifecycle events a stack reports. Shared so the legacy capability-URL route
// and the authenticated stack route accept exactly the same vocabulary.
const (
	PhoneHomeRequestTypeCreate = "Create"
	PhoneHomeRequestTypeUpdate = "Update"
	PhoneHomeRequestTypeDelete = "Delete"
)

// ValidPhoneHomeRequestType reports whether s is a known lifecycle event.
func ValidPhoneHomeRequestType(s string) bool {
	switch s {
	case PhoneHomeRequestTypeCreate, PhoneHomeRequestTypeUpdate, PhoneHomeRequestTypeDelete:
		return true
	default:
		return false
	}
}

// RecordStackPhoneHome persists a stack's reported outputs against a stack version:
// marks the version active and appends a run holding the payload. It returns the run
// so the caller can enqueue the stackrun signal that acts on it.
//
// Shared by both phone-home surfaces — the legacy public capability URL and the
// authenticated /v1/stacks/{install_id}/phone-home route — so they cannot drift.
// The enqueue stays with the callers because the stackrun signal package depends on
// this one, so importing it here would be a cycle.
func (h *Helpers) RecordStackPhoneHome(
	ctx context.Context,
	stackVersion *app.InstallStackVersion,
	req map[string]any,
) (*app.InstallStackVersionRun, error) {
	data, err := pkggenerics.ToMapstructureWithJSONTag(req)
	if err != nil {
		return nil, fmt.Errorf("unable to convert to mapstructure: %w", err)
	}
	hstoreData := generics.ToHstore(pkggenerics.ToStringMap(pkggenerics.EncodeNestedForHstore(data)))

	updatedStack := app.InstallStackVersion{
		ID: stackVersion.ID,
	}
	if res := h.db.WithContext(ctx).
		Model(&updatedStack).
		Updates(app.InstallStackVersion{
			Status: app.NewCompositeStatus(ctx, app.InstallStackVersionStatusActive),
			Runs: []app.InstallStackVersionRun{
				{
					Data: hstoreData,
				},
			},
		}); res.Error != nil {
		return nil, fmt.Errorf("unable to update stack version: %w", res.Error)
	}

	run := app.InstallStackVersionRun{
		OrgID:                 stackVersion.OrgID,
		CreatedByID:           stackVersion.CreatedByID,
		InstallStackVersionID: stackVersion.ID,
		Data:                  hstoreData,
	}
	if res := h.db.WithContext(ctx).Create(&run); res.Error != nil {
		return nil, fmt.Errorf("unable to create install stack version run: %w", res.Error)
	}

	return &run, nil
}
