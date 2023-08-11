package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateOrgRequest struct {
	CreatedByID string `json:"created_by_id,omitempty"`
	Name        string `json:"name,omitempty"`
}

func (s *service) CreateOrg(ctx *gin.Context) {
	req := CreateOrgRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.createOrg(ctx, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}

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
