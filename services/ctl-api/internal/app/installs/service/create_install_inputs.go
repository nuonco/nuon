package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	installupdated "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updated"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	pkgstate "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
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
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.InstallInputs
// @Router					/v1/installs/{install_id}/inputs [post]
func (s *service) CreateInstallInputs(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req CreateInstallInputsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
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

	// Pin to the install's app config, matching the inputs PATCH path. The app's
	// newest input config may belong to a config this install is not on.
	pinnedAppInputConfig, err := s.helpers.GetPinnedAppInputConfig(ctx, install.AppID, install.AppConfigID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get pinned app input config: %w", err))
		return
	}
	if pinnedAppInputConfig == nil || pinnedAppInputConfig.ID == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("no app input config on app config %s", install.AppConfigID),
			Description: "no app input configs defined",
		})
		return
	}

	if err := s.helpers.ValidateInstallInputs(ctx, pinnedAppInputConfig, req.Inputs); err != nil {
		ctx.Error(err)
		return
	}

	inputs, err := s.createInstallInputs(ctx, install, pinnedAppInputConfig, req.Inputs)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create install inputs: %w", err))
		return
	}

	queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
	if err != nil {
		ctx.Error(err)
		return
	}
	if err := s.enqueueInstallSignal(ctx, queueID, &installupdated.Signal{
		InstallID: install.ID,
	}, "", ""); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, inputs)
}

func (s *service) createInstallInputs(ctx context.Context, install *app.Install, appInputConfig *app.AppInputConfig, inputs map[string]*string) (*app.InstallInputs, error) {
	obj := &app.InstallInputs{
		AppInputConfigID: appInputConfig.ID,
		InstallID:        install.ID,
		Values:           pgtype.Hstore(inputs),
	}

	// under the lock so a migration cannot append a stale copy after this row
	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := helpers.LockInstallInputs(ctx, tx, install.ID); err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Create(&obj).Error; err != nil {
			return err
		}
		// stale_at alone is inert: the partial has to be named or state serves the old inputs
		return s.helpers.MarkInstallStatePartialsStale(ctx, tx, install.ID, pkgstate.PartialInputs)
	}); err != nil {
		return nil, fmt.Errorf("unable to create install inputs: %w", err)
	}

	return obj, nil
}
