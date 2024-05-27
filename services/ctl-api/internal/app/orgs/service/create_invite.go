package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type CreateOrgInviteRequest struct {
	Email string `json:"email"`
}

// @ID CreateOrgInvite
// @Summary	Invite a user to the current org
// @Description.markdown create_org_invite.md
// @Param			req	body	CreateOrgInviteRequest	true	"Input"
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.OrgInvite
// @Router			/v1/orgs/current/invites [POST]
func (s *service) CreateOrgInvite(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateOrgInviteRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	invite, err := s.createInvite(ctx, org.ID, req.Email)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create invite: %w", err))
		return
	}

	s.evClient.Send(ctx, org.ID, &sigs.Signal{
		Type:     sigs.OperationInviteCreated,
		InviteID: invite.ID,
	})
	ctx.JSON(http.StatusCreated, invite)
}

func (s *service) createInvite(ctx context.Context, orgID, email string) (*app.OrgInvite, error) {
	invite := app.OrgInvite{
		OrgID:  orgID,
		Email:  email,
		Status: app.OrgInviteStatusPending,
	}

	err := s.db.WithContext(ctx).
		Create(&invite).Error
	if err != nil {
		return nil, fmt.Errorf("unable to create invite: %w", err)
	}

	return &invite, nil
}
