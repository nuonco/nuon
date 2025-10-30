package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateInstallInputsRequest struct {
	Inputs map[string]*string `json:"inputs" validate:"required,gte=1"`
}

func (c *CreateInstallInputsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// DEPRECATED we should use the new UpdateInstallInputs

// @ID						CreateInstallInputs
// @Summary				create install inputs
// @Description.markdown	create_install_inputs.md
// @Tags					installs
// @Accept					json
// @Param					req	body	CreateInstallInputsRequest	true	"Input"
// @Produce				json
// @Param					install_id	path	string	true	"install ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.InstallInputs
// @Router					/v1/installs/{install_id}/inputs [post]
func (s *service) CreateInstallInputs(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallInputsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	install, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if len(install.App.AppInputConfigs) < 1 {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("no app input configs defined on app"),
			Description: "no app input configs defined",
		})
		return
	}

	latestAppInputConfig, err := s.helpers.GetLatestAppInputConfig(ctx, install.AppID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get latest app input config: %w", err))
		return
	}

	if err := s.helpers.ValidateInstallInputs(ctx, latestAppInputConfig, req.Inputs); err != nil {
		ctx.Error(err)
		return
	}

	inputs, err := s.createInstallInputs(ctx, install, req.Inputs)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install inputs: %w", err))
		return
	}

	s.evClient.Send(ctx, install.ID, &signals.Signal{
		Type: signals.OperationUpdated,
	})

	ctx.JSON(http.StatusCreated, inputs)
}

func (s *service) createInstallInputs(ctx context.Context, install *app.Install, inputs map[string]*string) (*app.InstallInputs, error) {
	obj := &app.InstallInputs{
		AppInputConfigID: install.App.AppInputConfigs[0].ID,
		InstallID:        install.ID,
		Values:           pgtype.Hstore(inputs),
	}
	res := s.db.WithContext(ctx).Create(&obj)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install inputs: %w", res.Error)
	}

	return obj, nil
}
