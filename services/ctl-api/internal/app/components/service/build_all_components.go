package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	buildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
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
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
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

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check features: %w", err))
		return
	}

	var blds []*app.ComponentBuild

	for _, c := range comp {
		bld, err := s.helpers.CreateComponentBuild(ctx, c.ID, true, nil)
		if err != nil {
			ctx.Error(err)
			return
		}

		if useQueues {
			q, err := s.queueClient.GetQueueByOwner(ctx, c.ID, "components")
			if err != nil {
				ctx.Error(fmt.Errorf("unable to get component queue: %w", err))
				return
			}

			if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
				QueueID:   q.ID,
				OwnerID:   bld.ID,
				OwnerType: "component_builds",
				Signal: &buildsignal.Signal{
					ComponentID: c.ID,
					BuildID:     bld.ID,
				},
			}); err != nil {
				ctx.Error(fmt.Errorf("unable to enqueue build signal: %w", err))
				return
			}
		} else {
			s.evClient.Send(ctx, c.ID, &signals.Signal{
				Type:    signals.OperationBuild,
				BuildID: bld.ID,
			})
		}

		blds = append(blds, bld)
	}

	ctx.JSON(http.StatusCreated, blds)
}
