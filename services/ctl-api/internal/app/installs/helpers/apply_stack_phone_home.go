package helpers

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// StackPhoneHomeRequest is the free-form payload an install stack reports: the
// stack's outputs plus a `request_type` naming the lifecycle event.
type StackPhoneHomeRequest map[string]any

// Lifecycle events a stack reports. The legacy route keeps its own copy.
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

// RecordStackPhoneHome marks the version active and appends a run. Returns the run
// for the caller to enqueue a signal against; doing it here would be an import cycle.
//
// Serves the authenticated route only — the legacy one keeps its own copy.
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

	var run app.InstallStackVersionRun
	var clearedAwait bool
	if err := h.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current app.InstallStackVersion
		if err := tx.First(&current, "id = ?", stackVersion.ID).Error; err != nil {
			return fmt.Errorf("unable to load stack version: %w", err)
		}
		wasCancelled := current.Status.Status == app.StatusCancelled

		// Drop the await hook so a cancelled provision workflow is not resumed.
		if wasCancelled {
			if res := tx.Model(&app.InstallStackVersion{}).
				Where("id = ?", stackVersion.ID).
				Update("callback_ref", callback.Ref{}); res.Error != nil {
				return fmt.Errorf("unable to clear await callback: %w", res.Error)
			}
			clearedAwait = true
		}

		updatedStack := app.InstallStackVersion{
			ID: stackVersion.ID,
		}
		if res := tx.Model(&updatedStack).
			Updates(app.InstallStackVersion{
				Status: app.NewCompositeStatus(ctx, app.InstallStackVersionStatusActive),
				Runs: []app.InstallStackVersionRun{
					{
						Data: hstoreData,
					},
				},
			}); res.Error != nil {
			return fmt.Errorf("unable to update stack version: %w", res.Error)
		}

		run = app.InstallStackVersionRun{
			OrgID:                 stackVersion.OrgID,
			CreatedByID:           stackVersion.CreatedByID,
			InstallStackVersionID: stackVersion.ID,
			Data:                  hstoreData,
		}
		if res := tx.Create(&run); res.Error != nil {
			return fmt.Errorf("unable to create install stack version run: %w", res.Error)
		}

		requestType, _ := req["request_type"].(string)
		if err := h.RecordInstallStackVersionApplied(ctx, tx, stackVersion, requestType); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}

	if clearedAwait {
		h.l.Info("cleared await callback before phone-home on cancelled stack version",
			zap.String("install_stack_version_id", stackVersion.ID),
			zap.String("install_id", stackVersion.InstallID),
		)
	}

	return &run, nil
}
