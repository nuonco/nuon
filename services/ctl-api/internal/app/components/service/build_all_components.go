package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/components/signals"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type BuildAllComponentsRequest struct{}

func (c *BuildAllComponentsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						BuildAllComponents
// @Summary				create component build
// @Description.markdown	build_all_components.md
// @Param					app_id	path	string						true	"component ID"
// @Param					req				body	BuildAllComponentsRequest	true	"Input"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{array}		app.ComponentBuild
// @Router					/v1/apps/{app_id}/components/build-all [POST]
func (s *service) BuildAllComponents(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req BuildAllComponentsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return

	}

	var comp []*app.Component
	limit := 10
	offset := 0

	for {
		var batch []*app.Component
		res := s.db.WithContext(ctx).
			Limit(limit).
			Offset(offset).
			Where("app_id = ?", appID).
			Find(&batch)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to list components: %w", res.Error))
			return
		}
		if len(batch) == 0 {
			break
		}
		comp = append(comp, batch...)
		offset += limit
	}

	var blds []*app.ComponentBuild

	for _, c := range comp {
		bld, err := s.helpers.CreateComponentBuild(ctx, c.ID, true, nil)
		if err != nil {
			ctx.Error(err)
			return
		}
		s.evClient.Send(ctx, c.ID, &signals.Signal{
			Type:    signals.OperationBuild,
			BuildID: bld.ID,
		})

		blds = append(blds, bld)
	}

	ctx.JSON(http.StatusCreated, blds)
}
