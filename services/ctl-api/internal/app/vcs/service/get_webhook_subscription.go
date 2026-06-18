package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetVCSConnectionWebhookSubscription
// @Summary				returns the webhook subscription for a vcs connection
// @Description.markdown	get_vcs_connection_webhook_subscription.md
// @Param					connection_id	path	string	true	"connection ID"
// @Tags					vcs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.VCSWebhookSubscription
// @Router					/v1/vcs/connections/{connection_id}/webhook-subscription [get]
func (s *service) GetWebhookSubscription(ctx *gin.Context) {
	connectionID := ctx.Param("connection_id")

	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	vcsConn, err := s.getConnection(ctx, currentOrg.ID, connectionID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get vcs connection: %w", err))
		return
	}

	var sub app.VCSWebhookSubscription
	res := s.db.WithContext(ctx).
		Where(app.VCSWebhookSubscription{GithubInstallID: vcsConn.GithubInstallID}).
		First(&sub)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get webhook subscription: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, sub)
}
