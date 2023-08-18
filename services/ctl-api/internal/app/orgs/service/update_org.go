package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type UpdateOrgRequest struct {
	Name string `json:"name"`
}

// @BasePath /v1/orgs/

// Update an org
// @Summary Update an org
// @Schemes
// @Description Update an org
// @Param org_id path string true "org ID for your current org"
// @Param req body UpdateOrgRequest true "Input"
// @Tags orgs
// @Accept json
// @Produce json
// @Success 201 {object} app.Org
// @Router /v1/orgs/{org_id} [PATCH]
func (s *service) UpdateOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req UpdateOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := s.updateOrg(ctx, orgID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusAccepted, org)
}

func (s *service) updateOrg(ctx context.Context, orgID string, req *UpdateOrgRequest) (*app.Org, error) {
	org := app.Org{
		ID: orgID,
	}

	err := s.db.WithContext(ctx).Model(&org).Updates(app.Org{Name: req.Name}).Error
	if err != nil {
		return nil, fmt.Errorf("unable to update org: %w", err)
	}

	return &org, nil
}
