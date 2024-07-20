package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz/permissions"
)

const (
	hostedInstallerServiceAccountEmailTemplate string = "%s-installers.nuon.co"
)

type CreateHostedInstallerServiceAccountRequest struct {
	Name string `json:"name"`
}

// @ID CreateHostedInstallerServiceAccount
// @Summary	create a service account for hosted installers
// @Description.markdown create_hosted_installer_service_account.md
// @Param			req	body	CreateHostedInstallerServiceAccountRequest	true	"Input"
// @Tags			installers/admin
// @Accept			json
// @Produce		json
// @Success		201	{array}	app.Account
// @Router			/v1/installers/hosted-installer-service-account [POST]
func (s *service) CreateHostedInstallerServiceAccount(ctx *gin.Context) {
	var req CreateHostedInstallerServiceAccountRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}

	acct, err := s.createHostedInstallerServiceAccount(ctx, req.Name)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create installer service account: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, acct)
}

func (s *service) createHostedInstallerServiceAccount(ctx context.Context, name string) (*app.Account, error) {
	email := fmt.Sprintf(hostedInstallerServiceAccountEmailTemplate, name)

	acct, err := s.authzClient.FindAccount(ctx, email)
	if err == nil {
		return acct, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	acct = &app.Account{
		Email:       email,
		Subject:     name,
		AccountType: app.AccountTypeCanary,
	}
	res := s.db.WithContext(ctx).
		Create(acct)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create account: %w", res.Error)
	}

	ctx = middlewares.SetRegContext(ctx, acct)

	// create role
	role := app.Role{
		// create admin role
		RoleType: app.RoleTypeHostedInstaller,
		Policies: []app.Policy{
			{
				Name: app.PolicyNameHostedInstaller,
				Permissions: pgtype.Hstore(map[string]*string{
					"*": permissions.PermissionAll.ToStrPtr(),
				}),
			},
		},
	}

	res = s.db.
		WithContext(ctx).
		Create(&role)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create role: %w", res.Error)
	}

	// attach role
	if err := s.authzClient.AddAccountRoleByID(ctx, role.ID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to create account role: %w", err)
	}

	return acct, nil
}
