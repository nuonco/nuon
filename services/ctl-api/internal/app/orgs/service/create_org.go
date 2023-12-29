package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/auth"
)

type CreateOrgRequest struct {
	Name string `json:"name" validate:"required"`

	// These fields are used to control the behaviour of the org.
	UseCustomCert  bool `json:"use_custom_cert"`
	UseSandboxMode bool `json:"use_sandbox_mode"`
}

func (c *CreateOrgRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID CreateOrg
// @Summary	create a new org
// @Description.markdown	create_org.md
// @Security APIKey
// @Param			req	body	CreateOrgRequest	true	"Input"
// @Tags			orgs
// @Accept			json
// @Produce		json
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	app.Org
// @Router			/v1/orgs [POST]
func (s *service) CreateOrg(ctx *gin.Context) {
	user, err := auth.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
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

	org, err := s.createOrg(ctx, user.Subject, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}

	s.hooks.Created(ctx, org.ID, org.SandboxMode)
	ctx.JSON(http.StatusCreated, org)
}

func (s *service) createOrg(ctx context.Context, userID string, req *CreateOrgRequest) (*app.Org, error) {
	org := app.Org{
		Name:              req.Name,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start and provision org",
		SandboxMode:       req.UseSandboxMode,
		CustomCert:        req.UseCustomCert,
	}
	if s.cfg.ForceSandboxMode {
		org.SandboxMode = true
	}

	if err := s.db.WithContext(ctx).Create(&org).Error; err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	userOrg := app.UserOrg{
		UserID: userID,
		OrgID:  org.ID,
	}
	if err := s.db.WithContext(ctx).Create(&userOrg).Error; err != nil {
		return nil, fmt.Errorf("unable to create user org: %w", err)
	}

	return &org, nil
}
