package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type AdminAddAppSandboxes struct{}

//		@BasePath	/v1/apps
//
//	 Add app sandboxes for all objects in the database
//
//		@Summary	add an app sandbox to all apps
//		@Schemes
//		@Description	update and app to the latest sandbox
//		@Param			req		body	AdminUpdateSandboxRequest	true	"Input"
//		@Tags			apps/admin
//		@Accept			json
//		@Produce		json
//		@Success		200	{boolean}	true
//		@Router			/v1/apps/admin-add-app-sandboxes [POST]
func (s *service) AdminAddAppSandboxes(ctx *gin.Context) {
	var req AdminAddAppSandboxes
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
		if err := s.adminCreateAppSandbox(ctx, app); err != nil {
			ctx.Error(fmt.Errorf("unable to create app sandbox: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) adminCreateAppSandbox(ctx context.Context, currentApp *app.App) error {
	ctx = context.WithValue(ctx, "user_id", currentApp.CreatedByID)
	ctx = context.WithValue(ctx, "org_id", currentApp.OrgID)

	// fetch latest built in sandbox
	var sandbox app.Sandbox
	res := s.db.WithContext(ctx).Preload("Releases", func(db *gorm.DB) *gorm.DB {
		return db.Order("sandbox_releases.created_at DESC")
	}).First(&sandbox, "name = ?", "aws-eks")
	if res.Error != nil {
		return fmt.Errorf("unable to get sandbox releases: %w", res.Error)
	}
	if len(sandbox.Releases) < 1 {
		return fmt.Errorf("no sandbox releases found for aws-eks: %w", res.Error)
	}

	// create app sandbox
	appSandbox := app.AppSandbox{
		AppID: currentApp.ID,
		AppSandboxConfigs: []app.AppSandboxConfig{
			{
				TerraformVersion: "v1.6.3",
				SandboxReleaseID: generics.ToPtr(sandbox.Releases[0].ID),
				SandboxInputs:    pgtype.Hstore(map[string]*string{}),
			},
		},
	}
	res = s.db.WithContext(ctx).Create(&appSandbox)
	if errors.Is(res.Error, gorm.ErrDuplicatedKey) {
		return nil
	}
	if res.Error != nil {
		return fmt.Errorf("unable to create app sandbox: %w", res.Error)
	}

	// update all installs for the app

	return nil
}
