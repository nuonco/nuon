package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type UpdateOrgRequest struct {
	Name string `json:"name" validate:"required"`
}

func (c *UpdateOrgRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/orgs/

// Update current org
// @Summary Update current org
// @Schemes
// @Description Update current org
// @Param req body UpdateOrgRequest true "Input"
// @Tags orgs
// @Accept json
// @Produce json
// @Success 200 {object} app.Org
// @Router /v1/orgs/current [PATCH]
func (s *service) UpdateOrg(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err = s.updateOrg(ctx, org.ID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *service) updateOrg(ctx context.Context, orgID string, req *UpdateOrgRequest) (*app.Org, error) {
	org := app.Org{
		ID: orgID,
	}
	res := s.db.WithContext(ctx).Model(&org).Updates(app.Org{
		Name: req.Name,
	})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return nil, fmt.Errorf("org not found")
	}

	return &org, nil
}
