package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	buildsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/build"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateComponentBuildRequest struct {
	GitRef    *string `validate:"required_unless=UseLatest true" json:"git_ref"`
	UseLatest bool    `validate:"required_without=GitRef" json:"use_latest"`
}

func (c *CreateComponentBuildRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateAppComponentBuild
// @Summary				create component build
// @Description.markdown	create_component_build.md
// @Param					app_id			path	string						true	"app ID"
// @Param					component_id	path	string						true	"component ID"
// @Param					req				body	CreateComponentBuildRequest	true	"Input"
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
// @Success				201	{object}	app.ComponentBuild
// @Router					/v1/apps/{app_id}/components/{component_id}/builds [POST]
func (s *service) CreateAppComponentBuild(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	cmpID := ctx.Param("component_id")
	cmp, err := s.getAppComponent(ctx, appID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app component: %w", err))
		return
	}

	var req CreateComponentBuildRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check features: %w", err))
		return
	}

	// When both AppBranches and Queues features are enabled, use queue-based path.
	if useQueues {
		bld, err := s.helpers.CreateComponentBuild(ctx, cmp.ID, req.UseLatest, req.GitRef)
		if err != nil {
			ctx.Error(err)
			return
		}

		q, err := s.queueClient.GetQueueByOwner(ctx, cmp.ID, "components")
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get component queue: %w", err))
			return
		}

		if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			OwnerID:   bld.ID,
			OwnerType: "component_builds",
			Signal: &buildsignal.Signal{
				ComponentID: cmp.ID,
				BuildID:     bld.ID,
			},
		}); err != nil {
			ctx.Error(fmt.Errorf("unable to enqueue build signal: %w", err))
			return
		}

		ctx.JSON(http.StatusCreated, bld)
		return
	}

	// Legacy path: direct event loop signal
	bld, err := s.helpers.CreateComponentBuild(ctx, cmp.ID, req.UseLatest, req.GitRef)
	if err != nil {
		ctx.Error(err)
		return
	}
	s.evClient.Send(ctx, cmpID, &signals.Signal{
		Type:    signals.OperationBuild,
		BuildID: bld.ID,
	})

	ctx.JSON(http.StatusCreated, bld)
}

// @ID						CreateComponentBuild
// @Summary				create component build
// @Description.markdown	create_component_build.md
// @Param					component_id	path	string						true	"component ID"
// @Param					req				body	CreateComponentBuildRequest	true	"Input"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.ComponentBuild
// @Router					/v1/components/{component_id}/builds [POST]
func (s *service) CreateComponentBuild(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	cmpID := ctx.Param("component_id")

	// Validate component belongs to org before creating build
	_, err = s.findComponent(ctx, org.ID, cmpID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find component %s: %w", cmpID, err))
		return
	}

	var req CreateComponentBuildRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check features: %w", err))
		return
	}

	bld, err := s.helpers.CreateComponentBuild(ctx, cmpID, req.UseLatest, req.GitRef)
	if err != nil {
		ctx.Error(err)
		return
	}

	if useQueues {
		q, err := s.queueClient.GetQueueByOwner(ctx, cmpID, "components")
		if err != nil {
			ctx.Error(fmt.Errorf("unable to get component queue: %w", err))
			return
		}

		if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			OwnerID:   bld.ID,
			OwnerType: "component_builds",
			Signal: &buildsignal.Signal{
				ComponentID: cmpID,
				BuildID:     bld.ID,
			},
		}); err != nil {
			ctx.Error(fmt.Errorf("unable to enqueue build signal: %w", err))
			return
		}
	} else {
		s.evClient.Send(ctx, cmpID, &signals.Signal{
			Type:    signals.OperationBuild,
			BuildID: bld.ID,
		})
	}

	ctx.JSON(http.StatusCreated, bld)
}
