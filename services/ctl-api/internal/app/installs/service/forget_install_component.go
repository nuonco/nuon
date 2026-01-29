package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ForgetInstallComponentRequest struct{}

// @ID						ForgetInstallComponent
// @Summary				forget an install component
// @Description.markdown	forget_install_component.md
// @Param					install_id		path	string							true	"install ID"
// @Param					component_id	path	string							true	"component ID"
// @Param					req				body	ForgetInstallComponentRequest	true	"Input"
// @Tags					installs
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{boolean}	true
// @Router					/v1/installs/{install_id}/components/{component_id}/forget [POST]
func (s *service) ForgetInstallComponent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")
	componentID := ctx.Param("component_id")

	// Verify install exists and belongs to org
	if _, err := s.findInstall(ctx, org.ID, installID); err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	// Get install component to verify it exists
	installComponent, err := s.getInstallComponent(ctx, installID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install component %s: %w", componentID, err))
		return
	}

	// Get install details to access app config ID
	install, err := s.findInstall(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install details: %w", err))
		return
	}

	// Validate component has been removed from app config
	activeComponentIDs, err := s.getAppConfigComponentIDs(ctx, install.AppConfigID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check app config: %w", err))
		return
	}

	if activeComponentIDs[componentID] {
		ctx.Error(stderr.ErrUser{
			Err:         errors.New("Component still in app config. Remove it before forgetting"),
			Description: "This component is still referenced in the app configuration. Please remove it from your app config by running 'nuon apps sync' after removing the component from your nuon.yaml file, then try again.",
		})
		return
	}

	// Perform the soft delete
	err = s.forgetInstallComponent(ctx, installComponent.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

// getAppConfigComponentIDs returns a map of component IDs currently in the app config
func (s *service) getAppConfigComponentIDs(ctx context.Context, appConfigID string) (map[string]bool, error) {
	var appConfig app.AppConfig
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigConnections").
		Where("id = ?", appConfigID).
		First(&appConfig)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}

	// Build map of active component IDs from ComponentConfigConnections (source of truth)
	componentIDs := make(map[string]bool)
	for _, ccc := range appConfig.ComponentConfigConnections {
		componentIDs[ccc.ComponentID] = true
	}

	return componentIDs, nil
}

// Private method to perform soft delete
func (s *service) forgetInstallComponent(ctx context.Context, installComponentID string) error {
	res := s.db.WithContext(ctx).Delete(&app.InstallComponent{
		ID: installComponentID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete install component: %w", res.Error)
	}

	if res.RowsAffected < 1 {
		return fmt.Errorf("install component not found %s %s", installComponentID, gorm.ErrRecordNotFound)
	}

	return nil
}
