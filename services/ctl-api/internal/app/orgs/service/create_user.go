package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateOrgUserRequest struct {
	UserID string `json:"user_id"`
}

// @ID						AddUser
// @Summary				Add a user to the current org
// @Description.markdown	create_org_user.md
// @Param					req	body	CreateOrgUserRequest	true	"Input"
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
// @Success				201	{object}	app.Account
// @Router					/v1/orgs/current/user [POST]
func (s *service) CreateUser(ctx *gin.Context) (require.PreparedMutation, error) {
	var req CreateOrgUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return require.PreparedMutation{}, stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		}
	}
	if req.UserID == "" {
		return require.PreparedMutation{}, stderr.NewInvalidRequest(fmt.Errorf("user_id is required"))
	}
	return require.PreparedMutation{SubjectAccountID: req.UserID, Mutate: s.createUser}, nil
}

func (s *service) createUser(ctx *gin.Context) (require.Response, error) {
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
		return require.Response{}, stderr.ErrUser{Err: fmt.Errorf("cannot change your own membership"), Description: "you cannot change your own membership"}
	}
	target, err := prepareOrgMember(tx, orgID, accountID, false)
	if err != nil {
		return require.Response{}, err
	}
	client := authz.New(authz.Params{DB: tx})
	role, err := client.ResolveDelegableRole(tx.Statement.Context, orgID, app.RoleTypeOrgReadOnly, app.RoleContextTeam)
	if err != nil {
		return require.Response{}, err
	}
	if err := client.AddAccountRoleAtScope(tx.Statement.Context, target.ID, role.ID, access.Target().Scope); err != nil {
		return require.Response{}, err
	}

	return require.Response{Status: http.StatusCreated, Body: target}, nil
}
