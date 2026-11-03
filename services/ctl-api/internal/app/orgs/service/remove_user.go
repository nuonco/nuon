package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type RemoveOrgUserRequest struct {
	UserID string `json:"user_id"`
}

// @ID						RemoveUser
// @Summary				Remove a user from the current org
// @Description.markdown	remove_org_user.md
// @Param					req	body	RemoveOrgUserRequest	true	"Input"
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
// @Success				201	{object}	app.Account
// @Router					/v1/orgs/current/remove-user [POST]
func (s *service) RemoveUser(ctx *gin.Context) (require.PreparedMutation, error) {
	var req RemoveOrgUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.PreparedMutation{}, stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		}
	}
	if req.UserID == "" {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(fmt.Errorf("user_id is required"))
	}
	return require.PreparedMutation{SubjectAccountID: req.UserID, Mutate: s.removeUser}, nil
}

func (s *service) removeUser(ctx *gin.Context) (require.Response, error) {
	access, err := require.MutationAccessFromContext(ctx)
	if err != nil {
		return require.Response{}, err
	}
	tx := access.DB()
	orgID := access.Target().OrgID
	accountID := access.SubjectAccountID()
	acct, err := cctx.AccountFromContext(tx.Statement.Context)
	if err != nil {
		return require.Response{}, err
	}
	if accountID == acct.ID {
		return require.Response{}, stderr.ErrUser{Err: fmt.Errorf("cannot remove yourself"), Description: "you cannot remove yourself from the org"}
	}
	target, err := prepareOrgMember(tx, orgID, accountID, true)
	if err != nil {
		return require.Response{}, err
	}
	if err := s.requireRemainingOrgAdmin(tx, orgID, target); err != nil {
		return require.Response{}, err
	}
	if err := authz.New(authz.Params{DB: tx}).RemoveAccountOrgRoles(tx.Statement.Context, orgID, accountID); err != nil {
		return require.Response{}, err
	}

	return require.Response{Status: http.StatusAccepted, Body: acct}, nil
}
