package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

const (
	// NOTE(jm): we eventually will allow an app to use a custom sandbox, but for now we just use aws-eks
	defaultSandboxName string = "aws-eks"
)

type CreateAppRequest struct {
	Name string `json:"name"`
}

// @BasePath /v1/apps
// Create an app
// @Summary create an app
// @Schemes
// @Description get an app
// @Param app_id path string app_id "app ID"
// @Tags apps
// @Accept json
// @Produce json
// @Success 201 {object} app.App
// @Router /v1/apps/ [post]
func (s *service) CreateApp(ctx *gin.Context) {
	var req CreateAppRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	app, err := s.createApp(ctx, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	s.hooks.Created(ctx, app.ID)
	ctx.JSON(http.StatusOK, app)
}

func (s *service) createApp(ctx context.Context, req *CreateAppRequest) (*app.App, error) {
	sandbox := app.Sandbox{
		Name: defaultSandboxName,
	}
	res := s.db.WithContext(ctx).Preload("Releases").First(&sandbox)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox: %w", res.Error)
	}

	if len(sandbox.Releases) < 1 {
		return nil, fmt.Errorf("at least one release must be created for sandbox %s", defaultSandboxName)
	}

	app := app.App{
		// TODO(jm): set these once we have properly figured out auth
		CreatedByID: "abc",
		OrgID:       "org6h27y0rsz1oocphdb7o54zh",

		Name:             req.Name,
		SandboxReleaseID: sandbox.Releases[0].ID,
	}

	res = s.db.WithContext(ctx).Create(&app)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create app: %w", res.Error)
	}

	return &app, nil
}
