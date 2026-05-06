package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
	componentdelete "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/v2/delete"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID						DeleteAppComponent
// @Summary				delete a component
// @Description.markdown	delete_component.md
// @Param					app_id			path	string	true	"app ID"
// @Param					component_id	path	string	true	"component ID"
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
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/apps/{app_id}/components/{component_id} [DELETE]
func (s *service) DeleteAppComponent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	componentID := ctx.Param("component_id")

	// Validate component belongs to org before deleting
	_, err = s.findComponent(ctx, org.ID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find component %s: %w", componentID, err))
		return
	}

	err = s.appsHelpers.DeleteAppComponent(ctx, componentID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.dispatchComponentDelete(ctx, componentID); err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

// @ID						DeleteComponent
// @Summary				delete a component
// @Description.markdown	delete_component.md
// @Param					component_id	path	string	true	"component ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated    true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/components/{component_id} [DELETE]
func (s *service) DeleteComponent(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	componentID := ctx.Param("component_id")

	// Validate component belongs to org before deleting
	_, err = s.findComponent(ctx, org.ID, componentID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find component %s: %w", componentID, err))
		return
	}

	err = s.appsHelpers.DeleteAppComponent(ctx, componentID)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.dispatchComponentDelete(ctx, componentID); err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

// dispatchComponentDelete enqueues the v2 delete signal in queue mode, otherwise
// falls back to the legacy event loop OperationDelete signal.
func (s *service) dispatchComponentDelete(ctx context.Context, componentID string) error {
	useQueues, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches, app.OrgFeatureQueues)
	if err != nil {
		return fmt.Errorf("unable to check features: %w", err)
	}

	if useQueues {
		q, err := s.queueClient.GetQueueByOwner(ctx, componentID, "components")
		if err != nil {
			return fmt.Errorf("unable to get component queue: %w", err)
		}
		if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			OwnerID:   componentID,
			OwnerType: "components",
			Signal:    &componentdelete.Signal{ComponentID: componentID},
		}); err != nil {
			return fmt.Errorf("unable to enqueue component delete signal: %w", err)
		}
		return nil
	}

	s.evClient.Send(ctx, componentID, &signals.Signal{
		Type: signals.OperationDelete,
	})
	return nil
}
