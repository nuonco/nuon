package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/analytics/events"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orginvitecreated "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/invite_created"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						ResendOrgInvite
// @Summary				Resend an org invite
// @Description.markdown	resend_org_invite.md
// @Param					invite_id	path	string	true	"invite ID"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.OrgInvite
// @Router					/v1/orgs/current/invites/{invite_id}/resend [POST]
func (s *service) ResendOrgInvite(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	inviteID := ctx.Param("invite_id")
	if inviteID == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invite_id is required"),
			Description: "invite_id is required",
		})
		return
	}

	invite, err := s.getOrgInviteByID(ctx, org.ID, inviteID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get invite: %w", err))
		return
	}

	if invite.Status != app.OrgInviteStatusPending {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invite has already been accepted"),
			Description: "invite has already been accepted",
		})
		return
	}

	queueID, err := s.getOrgSignalsQueueID(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org signals queue: %w", err))
		return
	}
	if err := s.enqueueOrgSignal(ctx, queueID, &orginvitecreated.Signal{
		OrgID:    org.ID,
		InviteID: invite.ID,
		LoginURL: fmt.Sprintf("%s/api/auth/login", s.cfg.AppURL),
	}, org.ID); err != nil {
		ctx.Error(fmt.Errorf("enqueue signal: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, invite)

	s.analyticsClient.Track(ctx, events.InviteResent, map[string]interface{}{
		"invite_id": invite.ID,
		"email":     invite.Email,
		"org_id":    invite.OrgID,
		"role_type": invite.RoleType,
	})
}

func (s *service) getOrgInviteByID(ctx *gin.Context, orgID, inviteID string) (*app.OrgInvite, error) {
	var invite app.OrgInvite

	res := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", inviteID, orgID).
		First(&invite)
	if res.Error != nil {
		return nil, res.Error
	}

	return &invite, nil
}
