package activities

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/deployerrors"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type RecordDeployBuildUnavailableCompositeErrorRequest struct {
	DeployID string `validate:"required"`

	Reason deployerrors.ComponentBuildUnavailableReason `validate:"required"`

	ComponentID   string
	ComponentName string

	BuildID                string
	BuildStatus            string
	BuildStatusDescription string
}

// RecordDeployBuildUnavailableCompositeError freezes a structured
// ComponentBuildUnavailableError onto the InstallDeploy when a deploy cannot
// proceed because its component build is unavailable (currently: the latest
// build is in a terminal failure state). This lets the dashboard guide the
// user to rebuild the component instead of surfacing a raw status string.
//
// It is best-effort: it does not block the deploy failure path.
//
// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) RecordDeployBuildUnavailableCompositeError(ctx context.Context, req RecordDeployBuildUnavailableCompositeErrorRequest) error {
	l := temporalzap.GetActivityLogger(ctx)
	l = l.With(
		zap.String("deploy_id", req.DeployID),
		zap.String("component_id", req.ComponentID),
		zap.String("build_id", req.BuildID),
	)

	ce := &deployerrors.ComponentBuildUnavailableError{
		Reason:                 req.Reason,
		ComponentID:            req.ComponentID,
		ComponentName:          req.ComponentName,
		BuildID:                req.BuildID,
		BuildStatus:            req.BuildStatus,
		BuildStatusDescription: req.BuildStatusDescription,
	}

	data := compositeerrors.New(ce)
	res := a.db.WithContext(ctx).
		Model(&app.InstallDeploy{ID: req.DeployID}).
		Select("composite_error").
		Updates(app.InstallDeploy{CompositeError: data})
	if res.Error != nil {
		return fmt.Errorf("unable to record deploy build-unavailable composite error: %w", res.Error)
	}
	if res.RowsAffected < 1 {
		return fmt.Errorf("no deploy found for id %s: %w", req.DeployID, gorm.ErrRecordNotFound)
	}

	l.Info("recorded build-unavailable composite error on deploy", zap.String("reason", string(req.Reason)))
	return nil
}
