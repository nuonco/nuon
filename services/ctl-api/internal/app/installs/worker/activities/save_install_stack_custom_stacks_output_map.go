package activities

import (
	"context"

	"github.com/pkg/errors"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SaveInstallStackVersionCustomStacksOutputMapRequest struct {
	ID                 string `validate:"required"`
	OutputMap          map[string]map[string]string
	InputParametersMap map[string]map[string]string
}

// SaveInstallStackVersionCustomStacksOutputMap persists the output and
// input-parameter mappings computed from the rendered custom-stacks-only
// template.
//
// @temporal-gen-v2 activity
func (a *Activities) SaveInstallStackVersionCustomStacksOutputMap(ctx context.Context, req *SaveInstallStackVersionCustomStacksOutputMapRequest) error {
	obj := &app.InstallStackVersion{ID: req.ID}

	res := a.db.WithContext(ctx).
		Model(&obj).Updates(app.InstallStackVersion{
		CustomStacksOutputMap:          req.OutputMap,
		CustomStacksInputParametersMap: req.InputParametersMap,
	})

	if res.Error != nil {
		return errors.Wrap(res.Error, "unable to update stack version custom stacks output map")
	}
	if res.RowsAffected != 1 {
		return errors.Wrap(gorm.ErrRecordNotFound, "install stack version not found")
	}

	return nil
}
