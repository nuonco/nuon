package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installID := ctx.Param("install_id")

	response, err := s.renderInstallReadme(ctx, org.ID, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	statusCode := http.StatusOK
	if len(response.Warnings) > 0 {
		statusCode = http.StatusPartialContent
	}

	ctx.JSON(statusCode, response)
}

func (s *service) renderInstallReadme(ctx context.Context, orgID, installID string) (*Readme, error) {
	install, err := s.helpers.GetInstall(ctx, orgID, installID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	installState, err := s.helpers.GetInstallState(ctx, install.ID, true, true)
	if err != nil {
		return nil, fmt.Errorf("unable to get install state: %w", err)
	}

	appConfig, err := s.appsHelpers.GetLatestActiveAppConfig(ctx, install.AppID)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest app config: %w", err)
	}

	stateMap, err := installState.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to json")
	}

	value, warnings, err := render.RenderWithWarnings(appConfig.Readme, stateMap)
	if err != nil {
		return nil, errors.Wrap(err, "unable to render")
	}
	return &Readme{
		Rendered: value,
		Original: appConfig.Readme,
		Warnings: generics.ErrsToStrings(warnings),
	}, nil
}
