package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/generics"
)

type AdminUpdateOrgsFeaturesRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

// @ID						AdminUpdateOrgsFeatures
// @Summary				update org features in bulk
// @Description.markdown	admin_update_orgs_features.md
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	AdminUpdateOrgsFeaturesRequest	true	"Input"
// @Produce				json
// @Success				200	{string}	ok
// @Router					/v1/orgs/admin-features  [PATCH]
func (s *service) AdminUpdateOrgsFeatures(ctx *gin.Context) {
	var req AdminUpdateOrgsFeaturesRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	err := s.bulkUpdateOrgFeatures(ctx, req.Features)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

func (s *service) bulkUpdateOrgFeatures(ctx context.Context, features map[string]bool) error {
	processBatch := func(orgs []*app.Org) error {
		for _, org := range orgs {
			err := s.features.Enable(ctx, org.ID, features)
			if err != nil {
				return fmt.Errorf("unable to update org features: %w", err)
			}
		}
		return nil
	}

	query := s.db.Model(&app.Org{}).Order("created_at ASC")
	return generics.BatchProcessing(ctx, 50, query, processBatch)
}
