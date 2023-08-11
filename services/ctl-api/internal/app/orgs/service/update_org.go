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

func (s *service) UpdateOrg(ctx *gin.Context) {
	orgID := ctx.Param("id")
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
		Model: app.Model{
			ID: orgID,
		},
	}

	err := s.db.WithContext(ctx).Model(&org).Updates(app.Org{Name: req.Name}).Error
	if err != nil {
		return nil, fmt.Errorf("unable to update org: %w", err)
	}

	return &org, nil
}
