package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminAddAppSandboxConfigs struct{}

// @BasePath	/v1/apps
//
//	Add app inputs for all apps
//
// @Summary	add inputs for all apps
// @Schemes
// @Description	add app inputs for all apps
// @Param			req	body	AdminAddAppSandboxConfigs	true	"Input"
// @Tags			apps/admin
// @Accept			json
// @Produce		json
// @Success		200	{boolean}	true
// @Router			/v1/apps/admin-add-app-sandbox-configs [POST]
func (s *service) AdminAddAppSandboxConfigs(ctx *gin.Context) {
	var req AdminAddAppSandboxConfigs
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	apps, err := s.getAllApps(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to fetch apps: %w", err))
		return
	}

	for _, app := range apps {
		if len(app.AppSandboxConfigs) > 0 {
			continue
		}

		if err := s.adminCreateAppSandboxConfig(ctx, app); err != nil {
			ctx.Error(fmt.Errorf("unablet to create sandbox config: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getLatestSandboxRelease(ctx context.Context, sandboxName string) (*app.SandboxRelease, error) {
	sandbox := app.Sandbox{}
	res := s.db.WithContext(ctx).
		First(&sandbox, "name = ?", sandboxName)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox: %w", res.Error)
	}

	release := app.SandboxRelease{}
	res = s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(1).
		First(&release, "sandbox_id = ?", sandbox.ID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox releases: %w", res.Error)
	}

	return &release, nil
}

func (s *service) adminCreateAppSandboxConfig(ctx context.Context, currentApp *app.App) error {
	sandboxRelease, err := s.getLatestSandboxRelease(ctx, "aws-eks")
	if err != nil {
		return fmt.Errorf("unable to get latest sandbox release: %w", err)
	}

	appSandboxConfigs := app.AppSandboxConfig{
		CreatedByID:      currentApp.CreatedByID,
		OrgID:            currentApp.OrgID,
		AppID:            currentApp.ID,
		TerraformVersion: "v1.6.3",
		SandboxReleaseID: generics.ToPtr(sandboxRelease.ID),
	}

	res := s.db.WithContext(ctx).Create(&appSandboxConfigs)
	if res.Error != nil {
		return fmt.Errorf("unable to create app inputs %w", res.Error)
	}

	return nil
}
