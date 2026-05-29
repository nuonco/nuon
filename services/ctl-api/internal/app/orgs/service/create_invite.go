package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/analytics/events"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	sigs "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals"
	orginvitecreated "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/v2/invite_created"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateOrgInviteRequest struct {
	Email    string       `json:"email" validate:"required"`
	RoleType app.RoleType `json:"role_type,omitempty"`
}

// @ID						CreateOrgInvite
// @Summary				Invite a user to the current org
// @Description.markdown	create_org_invite.md
// @Param					req	body	CreateOrgInviteRequest	true	"Input"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.OrgInvite
// @Router					/v1/orgs/current/invites [POST]
func (s *service) CreateOrgInvite(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateOrgInviteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}

	if req.Email == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("email is required"),
			Description: "email is required",
		})
		return
	}

	if !helpers.IsEmail(req.Email) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid email"),
			Description: "invalid email",
		})

		return
	}

	roleType := req.RoleType
	if roleType == "" {
		roleType = app.RoleTypeOrgAdmin
	}
	if roleType != app.RoleTypeOrgAdmin && roleType != app.RoleTypeOrgSupport {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid role type: %s", roleType),
			Description: fmt.Sprintf("role_type must be %q or %q", app.RoleTypeOrgAdmin, app.RoleTypeOrgSupport),
		})
		return
	}

	invite, err := s.createInvite(ctx, org.ID, req.Email, roleType)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create invite: %w", err))
		return
	}

	useQueues, err := s.useOrgQueues(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("checking features: %w", err))
		return
	}
	if useQueues {
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
	} else {
		s.evClient.Send(ctx, org.ID, &sigs.Signal{
			Type:     sigs.OperationInviteCreated,
			InviteID: invite.ID,
		})
	}
	ctx.JSON(http.StatusCreated, invite)

	s.analyticsClient.Track(ctx, events.InviteSent, map[string]interface{}{
		"invite_id": invite.ID,
		"email":     invite.Email,
		"org_id":    invite.OrgID,
		"role_type": invite.RoleType,
	})
}

func (s *service) createInvite(ctx context.Context, orgID, email string, roleType app.RoleType) (*app.OrgInvite, error) {
	invite := app.OrgInvite{
		OrgID:    orgID,
		Email:    email,
		Status:   app.OrgInviteStatusPending,
		RoleType: roleType,
	}

	err := s.db.WithContext(ctx).
		Create(&invite).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create invite: %w", err)
	}

	return &invite, nil
}
