package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/account"
)

type AdminCreateServiceAccountRequest struct{}

//	@ID						AdminCreateServiceAccount
//	@Summary				create a service account for an org
//	@Description.markdown	create_org_service_account.md
//	@Param					req		body	AdminCreateServiceAccountRequest	true	"Input"
//	@Param					org_id	path	string								true	"org ID or name"
//	@Security				AdminEmail
//	@Tags					orgs/admin
//	@Accept					json
//	@Produce				json
//	@Success				201	{object}	app.Account
//	@Router					/v1/orgs/{org_id}/admin-service-account [POST]
func (s *service) AdminCreateServiceAccount(ctx *gin.Context) {
	var req AdminCreateServiceAccountRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
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
	email := account.ServiceAccountEmail(name)

	acct, err := s.acctClient.FindAccount(ctx, email)
	if err == nil {
		return acct, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to lookup account: %w", err)
	}

	newAcct := app.Account{
		Email:       email,
		Subject:     name,
		AccountType: app.AccountTypeService,
	}
	res := s.db.WithContext(ctx).
		Create(&newAcct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create service account: %w", res.Error)
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, orgID, newAcct.ID); err != nil {
		return nil, fmt.Errorf("unable to add org role to service account: %w", err)
	}

	return &newAcct, nil
}
