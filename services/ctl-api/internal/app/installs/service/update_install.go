package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/helpers"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/patcher"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
	"gorm.io/gorm"
)

type UpdateInstallRequest struct {
	Name     string                   `json:"name"`
	Metadata *helpers.InstallMetadata `json:"metadata,omitempty"`
}

func (c *UpdateInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						UpdateInstall
// @Summary				update an install
// @Description.markdown	update_install.md
// @Param					install_id	path	string					true	"app ID"
// @Param					req			body	UpdateInstallRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Install
// @Router					/v1/installs/{install_id} [PATCH]
func (s *service) UpdateInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req UpdateInstallRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse update request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.updateInstall(ctx, installID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}

func (s *service) updateInstall(ctx context.Context, installID string, req *UpdateInstallRequest) (*app.Install, error) {
	currentInstall := app.Install{
		ID: installID,
	}

	updateObj := app.Install{Name: req.Name}
	if req.Metadata != nil {
		updateObj.Metadata = generics.ToHstore(map[string]string{
			"managed_by": req.Metadata.ManagedBy,
		})
	}

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithPatcher(patcher.PatcherOptions{})).
		Model(&currentInstall).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("AppSandboxConfig").
		UpdateColumns(&updateObj)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("install not found: %w", gorm.ErrRecordNotFound)
	}

	return &currentInstall, nil
}
