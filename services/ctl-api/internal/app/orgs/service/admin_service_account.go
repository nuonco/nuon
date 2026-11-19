package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminCreateServiceAccountRequest struct{}

// @ID						AdminCreateServiceAccount
// @Summary				create a service account for an org
// @Description.markdown	admin_create_org_service_account.md
// @Param					req		body	AdminCreateServiceAccountRequest	true	"Input"
// @Param					org_id	path	string								true	"org ID or name"
// @Security				AdminEmail
// @Tags					orgs/admin
// @Accept					json
// @Produce				json
// @Success				201	{object}	app.Account
// @Router					/v1/orgs/{org_id}/admin-service-account [POST]
func (s *service) AdminCreateServiceAccount(ctx *gin.Context) {
	var req AdminCreateServiceAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	orgID := ctx.Param("org_id")
	org, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	acct, err := s.createServiceAccount(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create account: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, acct)
}

func (s *service) createServiceAccount(ctx context.Context, orgID string) (*app.Account, error) {
	name := fmt.Sprintf("%s-admin-service-account", orgID)
	return s.acctClient.CreateOrgServiceAccount(ctx, orgID, name, "", app.RoleTypeOrgAdmin)
}
