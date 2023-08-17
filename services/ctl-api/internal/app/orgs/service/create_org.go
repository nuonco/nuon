package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateOrgRequest struct {
	CreatedByID string `json:"created_by_id,omitempty"`
	Name        string `json:"name,omitempty"`
}

func (c *CreateOrgRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/orgs

// Create a new org
// @Summary create a new org
// @Schemes
// @Description create a new org
// @Param req body CreateOrgRequest true "Input"
// @Tags orgs
// @Accept json
// @Produce json
// @Success 201 {object} app.Org
// @Router /v1/orgs [POST]
func (s *service) CreateOrg(ctx *gin.Context) {
	req := CreateOrgRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := s.createOrg(ctx, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}

	s.hooks.Created(ctx, org.ID)
	ctx.JSON(http.StatusCreated, org)
}

func (s *service) createOrg(ctx context.Context, req *CreateOrgRequest) (*app.Org, error) {
	org := &app.Org{
		CreatedByID: req.CreatedByID,
		Name:        req.Name,
	}
	if err := s.db.WithContext(ctx).Create(org).Error; err != nil {
		return nil, fmt.Errorf("unable to create org: %w", err)
	}

	return org, nil
}
