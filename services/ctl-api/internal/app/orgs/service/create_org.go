package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	sigs "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type CreateOrgRequest struct {
	Name string `json:"name" validate:"required"`

	// These fields are used to control the behaviour of the org.
	UseSandboxMode bool `json:"use_sandbox_mode"`
}

func (c *CreateOrgRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID						CreateOrg
// @Summary				create a new org
// @Description.markdown	create_org.md
// @Security				APIKey
// @Param					req	body	CreateOrgRequest	true	"Input"
// @Tags					orgs
// @Accept					json
// @Produce				json
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.Org
// @Router					/v1/orgs [POST]
func (s *service) CreateOrg(ctx *gin.Context) {
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	// NOTE(jm): we have a discrepancy in your db, where some OG members had their data inputted into the accounts
	// table backwards.
	isAllowed := false
	for _, allowed := range strings.Split(s.cfg.OrgCreationEmailAllowList, ",") {
		if strings.HasSuffix(acct.Email, allowed) || strings.HasSuffix(acct.Subject, allowed) {
			isAllowed = true
		}
	}

	if acct.AccountType == app.AccountTypeAuth0 && !isAllowed {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("This email is not allowed to create new orgs."),
			Description: "Please reach out to team@nuon.co for access.",
		})
		return
	}

	req := CreateOrgRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	newOrg, err := s.createOrg(ctx, acct, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}
	cctx.SetOrgGinContext(ctx, newOrg)

	s.evClient.Send(ctx, newOrg.ID, &sigs.Signal{
		Type: sigs.OperationCreated,
	})
	s.evClient.Send(ctx, newOrg.ID, &sigs.Signal{
		Type: sigs.OperationProvision,
	})

	ctx.JSON(http.StatusCreated, newOrg)
}

func (s *service) createOrg(ctx context.Context, acct *app.Account, req *CreateOrgRequest) (*app.Org, error) {
	orgTyp := app.OrgTypeDefault
	if req.UseSandboxMode {
		orgTyp = app.OrgTypeSandbox
	}
	if acct.AccountType == app.AccountTypeIntegration {
		orgTyp = app.OrgTypeIntegration
	}
	if s.cfg.ForceSandboxMode {
		orgTyp = app.OrgTypeSandbox
	}

	notificationsCfg := app.NotificationsConfig{
		EnableSlackNotifications: acct.AccountType == app.AccountTypeAuth0,
		EnableEmailNotifications: acct.AccountType == app.AccountTypeAuth0,
		InternalSlackWebhookURL:  s.cfg.InternalSlackWebhookURL,
	}
	org := app.Org{
		Name:                req.Name,
		Status:              "queued",
		StatusDescription:   "waiting for event loop to start and provision org",
		SandboxMode:         req.UseSandboxMode,
		OrgType:             orgTyp,
		NotificationsConfig: notificationsCfg,
	}
	if s.cfg.ForceSandboxMode {
		org.SandboxMode = true
	}
	if s.cfg.ForceDebugMode {
		org.DebugMode = true
	}

	if err := s.db.WithContext(ctx).Create(&org).Error; err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	// make sure the notifications config orgID is set
	if res := s.db.WithContext(ctx).
		Where(&app.NotificationsConfig{
			OwnerID: org.ID,
		}).
		Updates(app.NotificationsConfig{
			OrgID: org.ID,
		}); res.Error != nil {
		return nil, fmt.Errorf("unable to set org ID on notifications config: %w", res.Error)
	}

	if err := s.authzClient.CreateOrgRoles(ctx, org.ID); err != nil {
		return nil, fmt.Errorf("unable to create org roles: %w", err)
	}

	if err := s.authzClient.AddAccountOrgRole(ctx, app.RoleTypeOrgAdmin, org.ID, acct.ID); err != nil {
		return nil, fmt.Errorf("unable to add user to org: %w", err)
	}

	if _, err := s.runnersHelpers.CreateOrgRunnerGroup(ctx, &org); err != nil {
		return nil, fmt.Errorf("unable to create org runner group: %w", err)
	}

	return &org, nil
}
