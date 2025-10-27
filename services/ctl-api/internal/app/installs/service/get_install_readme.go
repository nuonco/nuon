package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type Readme struct {
	Original string   `json:"original"`
	Warnings []string `json:"warnings"`

	Rendered string `json:"readme"`
}

// @ID						GetInstallReadme
// @Summary				get install readme rendered with
// @Description.markdown	get_install_readme.md
// @Param					install_id	path	string	true	"install ID"
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
// @Success				200	{object}	Readme
// @Success				206	{object}	Readme
// @Router					/v1/installs/{install_id}/readme [get]
func (s *service) GetInstallReadme(ctx *gin.Context) {
	// get install state
	installID := ctx.Param("install_id")

	install, err := s.helpers.GetInstall(ctx, installID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get install"))
		return
	}

	installState, err := s.helpers.GetInstallState(ctx, installID, true, true)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install state: %w", err))
		return
	}

	// get app readme template
	appConfig, err := s.appsHelpers.GetAppLatestConfig(ctx, install.AppID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest app config: %w", err))
		return
	}

	// interpolate the state into the readme md
	stateMap, err := installState.AsMap()
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to convert state to json"))
		return
	}

	value, warnings, err := render.RenderWithWarnings(appConfig.Readme, stateMap)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to render"))
		return
	}

	response := Readme{
		Rendered: value,
		Original: appConfig.Readme,
		Warnings: generics.ErrsToStrings(warnings),
	}

	statusCode := http.StatusOK
	if len(warnings) > 0 {
		statusCode = http.StatusPartialContent
	}

	ctx.JSON(statusCode, response)
}

func (s *service) getLatestAppConfig(ctx context.Context, appID string) (*app.AppConfig, error) {
	var appConfig app.AppConfig
	res := s.db.WithContext(ctx).Where("app_id = ?", appID).Order("created_at DESC").First(&appConfig)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app config: %w", res.Error)
	}
	return &appConfig, nil
}
