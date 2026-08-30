package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/stackerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type SetInstallStackVersionCompositeErrorRequest struct {
	StackVersionID string `validate:"required"`
	// Platform is the cloud target ("aws", "azure", "gcp") for the error message.
	Platform string
	// Detail is the sanitised renderer error message.
	Detail string
}

// SetInstallStackVersionCompositeError freezes a StackTemplateRenderError onto
// the given stack version row. Passing an empty Detail clears the field (nil),
// which is used at the start of a retry to remove a stale error.
//
// @temporal-gen-v2 activity
// @max-retries 3
func (a *Activities) SetInstallStackVersionCompositeError(ctx context.Context, req SetInstallStackVersionCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx).With(zap.String("stack_version_id", req.StackVersionID))

	var data *compositeerrors.CompositeErrorData
	if req.Detail != "" {
		var err error
		data, err = compositeerrors.New(
			&stackerrors.StackTemplateRenderError{
				Platform: req.Platform,
				Detail:   req.Detail,
			},
			compositeerrors.WithSource("install_stack_versions", req.StackVersionID),
		)
		if err != nil {
			return fmt.Errorf("unable to build stack template render composite error: %w", err)
		}
	}

	res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{ID: req.StackVersionID}).
		Select("composite_error").
		Updates(app.InstallStackVersion{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to set install stack version composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no stack version found for id %s: %w", req.StackVersionID, gorm.ErrRecordNotFound)
	}

	l.Info("updated install stack version composite error")
	return nil
}
