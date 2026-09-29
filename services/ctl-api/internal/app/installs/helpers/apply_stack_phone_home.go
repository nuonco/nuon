package helpers

import (
	"context"
	"fmt"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

type StackPhoneHomeRequest map[string]any

const (
	PhoneHomeRequestTypeCreate = "Create"
	PhoneHomeRequestTypeUpdate = "Update"
	PhoneHomeRequestTypeDelete = "Delete"
)

func ValidPhoneHomeRequestType(s string) bool {
	switch s {
	case PhoneHomeRequestTypeCreate, PhoneHomeRequestTypeUpdate, PhoneHomeRequestTypeDelete:
		return true
	default:
		return false
	}
}

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

	requestType, _ := req["request_type"].(string)
	if err := h.RecordInstallStackVersionApplied(ctx, stackVersion, requestType); err != nil {
		return nil, err
	}

	return &run, nil
}
