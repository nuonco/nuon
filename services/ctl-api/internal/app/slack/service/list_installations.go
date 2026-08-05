package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ListSlackInstallations
// @Summary				List Slack workspaces linked to the current org
// @Description			Returns every active Slack workspace installation that has a verified org link to the calling org.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string	true	"Org ID"
// @Success				200	{array}		app.SlackInstallation
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/installations [GET]
func (s *service) ListInstallations(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	installs, err := s.listOrgInstallations(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to list slack installations: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

// Gated by ownership, not a verified link: the auto-link gives non-owner orgs links they must not act on.
func (s *service) listOrgInstallations(ctx context.Context, orgID string) ([]app.SlackInstallation, error) {
	var installs []app.SlackInstallation

	res := s.db.WithContext(ctx).
		Where(app.SlackInstallation{
			OwnerOrgID: orgID,
			Status:     app.SlackInstallationStatusActive,
		}).
		Order("slack_installations.created_at DESC").
		Find(&installs)
	if res.Error != nil {
		return nil, res.Error
	}
	return installs, nil
}
