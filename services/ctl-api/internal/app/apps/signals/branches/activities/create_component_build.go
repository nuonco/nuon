package activities

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateComponentBuildRequest struct {
	BuildID     string `validate:"required"`
	ComponentID string `validate:"required"`
	WorkflowID  string `validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateComponentBuild(ctx context.Context, req CreateComponentBuildRequest) (*app.ComponentBuild, error) {
	var build *app.ComponentBuild
	decision, err := a.authzClient.AdmitWorkflow(ctx, req.WorkflowID, func(tx *gorm.DB, wf *app.Workflow) error {
		if wf.Type != app.WorkflowTypeAppConfigBuild || wf.OwnerType != "apps" || wf.Request == nil || wf.Request.PinnedAppConfigID == "" {
			return temporal.NewNonRetryableApplicationError("build requires an accepted app-config build request", authz.WorkflowAuthorizationDenied, nil)
		}
		var config app.AppConfig
		if err := tx.Where(app.AppConfig{ID: wf.Request.PinnedAppConfigID, AppID: wf.OwnerID, OrgID: wf.OrgID}).Take(&config).Error; err != nil {
			return err
		}
		var component app.Component
		if err := tx.Where(app.Component{ID: req.ComponentID, AppID: wf.OwnerID, OrgID: wf.OrgID}).Take(&component).Error; err != nil {
			return err
		}
		var connection app.ComponentConfigConnection
		if err := tx.Where(app.ComponentConfigConnection{AppConfigID: config.ID, ComponentID: component.ID, OrgID: wf.OrgID}).Take(&connection).Error; err != nil {
			return err
		}
		var existing app.ComponentBuild
		err := tx.Where(app.ComponentBuild{ID: req.BuildID}).Take(&existing).Error
		if err == nil {
			if existing.OrgID != wf.OrgID || existing.ComponentConfigConnectionID != connection.ID || existing.CreatedByID != wf.CreatedByID {
				return temporal.NewNonRetryableApplicationError("build does not match the accepted request", authz.WorkflowAuthorizationDenied, nil)
			}
			build = &existing
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		buildCtx := cctx.SetOrgIDContext(cctx.SetAccountIDContext(ctx, wf.CreatedByID), wf.OrgID)
		build, err = a.componentHelpers.CreateComponentBuildWithID(buildCtx, tx, req.BuildID, component.ID, connection.ID)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("unable to admit component build: %w", err)
	}
	if !decision.Allowed {
		return nil, temporal.NewNonRetryableApplicationError("workflow authorization failed; submit a new request", authz.WorkflowAuthorizationDenied, nil, decision.Reason)
	}
	return build, nil
}
