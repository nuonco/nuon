package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	slackautolink "github.com/nuonco/nuon/services/ctl-api/internal/app/general/signals/v2/slack_auto_link"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

// @ID				AdminSlackAutoLink
// @Summary		Trigger Slack auto-link reconciliation
// @Description	Enqueues a one-shot general-slack-auto-link signal that reconciles SlackOrgLink rows for orgs matching the configured label gate, and (if SLACK_AUTO_LINK_CHANNEL_ID is set) seeds a default org-wide subscription per link. Idempotent and additive — never removes existing links or subs.
// @Tags			general/admin
// @Security		AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{object}	app.EmptyResponse
// @Router			/v1/general/slack-auto-link [POST]
func (s *service) AdminSlackAutoLink(ctx *gin.Context) {
	q, err := s.generalHelpers.EnsureGeneralQueue(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to ensure general queue: %w", err))
		return
	}

	if _, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal:  &slackautolink.Signal{},
	}); err != nil {
		ctx.Error(fmt.Errorf("unable to enqueue slack auto-link signal: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
