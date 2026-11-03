package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/analytics/events"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	orghelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	orginvitecreated "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/signals/invite_created"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CreateOrgInviteRequest struct {
	Email    string             `json:"email" validate:"required"`
	RoleType string             `json:"role_type,omitempty"`
	Scope    *permissions.Scope `json:"scope,omitempty"`
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
func (s *service) CreateOrgInvite(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}

	var req CreateOrgInviteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		}
	}

	if req.Email == "" {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("email is required"),
			Description: "email is required",
		}
	}

	if !helpers.IsEmail(req.Email) {
		return require.Response{}, stderr.ErrUser{
			Err:         fmt.Errorf("invalid email"),
			Description: "invalid email",
		}
	}

	roleType := app.RoleType(req.RoleType)
	if roleType == "" {
		roleType = app.RoleTypeOrgReadOnly
	}

	orgID := access.Target().OrgID
	scope, err := authz.RoleBindingScope(orgID, req.Scope)
	if err != nil {
		return require.Response{}, stderr.NewInvalidRequest(err)
	}
	tx := access.DB()
	invite, err := s.createInviteInTransaction(tx, orgID, req.Email, roleType, scope)
	if err != nil {
		return require.Response{}, err
	}
	signalID, err := s.enqueueInviteInTransaction(tx, invite)
	if err != nil {
		return require.Response{}, err
	}
	ctx.Set("org-invite-mutation", inviteMutation{Invite: *invite, SignalID: signalID, Event: events.InviteSent})
	return require.Response{Status: http.StatusCreated, Body: invite}, nil
}

type inviteMutation struct {
	Invite   app.OrgInvite
	SignalID string
	Event    events.Event
}

func (s *service) finishInviteMutation(ctx *gin.Context) {
	ctx.Next()
	if ctx.IsAborted() || len(ctx.Errors) != 0 || ctx.Writer.Status() >= http.StatusBadRequest {
		return
	}
	value, ok := ctx.Get("org-invite-mutation")
	if !ok {
		return
	}
	mutation := value.(inviteMutation)
	if mutation.SignalID != "" {
		s.queueClient.NotifySignal(mutation.SignalID)
	}
	s.analyticsClient.Track(ctx, mutation.Event, map[string]interface{}{
		"invite_id": mutation.Invite.ID,
		"email":     mutation.Invite.Email,
		"org_id":    mutation.Invite.OrgID,
		"role_type": mutation.Invite.RoleType,
	})
}

func (s *service) createInviteInTransaction(tx *gorm.DB, orgID, email string, roleType app.RoleType, scope permissions.Scope) (*app.OrgInvite, error) {
	invite := app.OrgInvite{
		OrgID:    orgID,
		Email:    email,
		Status:   app.OrgInviteStatusPending,
		RoleType: roleType,
		Scope:    scope,
	}

	if err := authz.New(authz.Params{DB: tx}).AuthorizeInvite(tx.Statement.Context, &invite); err != nil {
		return nil, err
	}
	if err := tx.Create(&invite).Error; err != nil {
		return nil, fmt.Errorf("unable to create invite: %w", err)
	}

	return &invite, nil
}

func (s *service) enqueueInviteInTransaction(tx *gorm.DB, invite *app.OrgInvite) (string, error) {
	var queue app.Queue
	if err := tx.Where(app.Queue{OwnerID: invite.OrgID, Name: orghelpers.OrgSignalsQueueName}).First(&queue).Error; err != nil {
		return "", err
	}
	response, err := s.queueClient.EnqueueSignalInTransaction(tx.Statement.Context, tx, &queueclient.EnqueueSignalRequest{
		QueueID:   queue.ID,
		Signal:    &orginvitecreated.Signal{OrgID: invite.OrgID, InviteID: invite.ID, LoginURL: s.cfg.AppURL},
		OwnerID:   invite.OrgID,
		OwnerType: plugins.TableName(tx, app.Org{}),
	})
	if err != nil {
		return "", err
	}
	return response.ID, nil
}
