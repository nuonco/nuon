package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminUpdateOrgsFeaturesRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

// @ID AdminUpdateOrgsFeatures
// @Summary get available org features
// @Description.markdown admin_update_orgs_features.md
// @Tags			orgs/admin
// @Security AdminEmail
// @Accept			json
// @Param			req	body AdminUpdateOrgsFeaturesRequest	true	"Input"
// @Produce		json
// @Success		200 {string}	ok
// @Router			/v1/orgs/admin-features  [PATCH]
func (s *service) AdminUpdateOrgsFeatures(ctx *gin.Context) {
	var req AdminUpdateOrgsFeaturesRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	err := s.validateOrgFeatures(ctx, req.Features)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	err = s.bulkUpdateOrgFeatures(ctx, req.Features)
	if err != nil {
		ctx.Error(fmt.Errorf("unable update org: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "ok")
}

func (s *service) bulkUpdateOrgFeatures(ctx context.Context, features map[string]bool) error {
	batchSize := 50
	var orgs []*app.Org
	offset := 0

	for {
		result := s.db.
			Offset(offset).
			Limit(batchSize).
			Find(&orgs).
			Order("created_at ASC")

		if result.Error != nil {
			return fmt.Errorf("unable to fetch orgs: %w", result.Error)
		}

		if len(orgs) == 0 {
			break
		}

		for _, org := range orgs {
			err := s.updateOrgFeatures(ctx, org, features)
			if err != nil {
				return fmt.Errorf("unable to update org features: %w", err)
			}
		}

		offset += batchSize
	}

	return nil
}
