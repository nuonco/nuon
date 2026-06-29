package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminToggleOrgsFeatureRequest struct {
	Features map[string]bool `json:"features" validate:"required"`
}

// @ID			AdminToggleOrgsFeature
// @Summary		toggle feature flags for all orgs
// @Description	Patches the provided feature flags (to their on/off value) onto every org via a jsonb merge, preserving each org's existing flags.
// @Tags			orgs/admin
// @Security		AdminEmail
// @Accept			json
// @Param			req	body	AdminToggleOrgsFeatureRequest	true	"Input"
// @Produce		json
// @Success		200	{object}	app.EmptyResponse
// @Router			/v1/orgs/admin-toggle-feature  [PATCH]
func (s *service) AdminToggleOrgsFeature(ctx *gin.Context) {
	var req AdminToggleOrgsFeatureRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if len(req.Features) == 0 {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("at least one feature is required")))
		return
	}

	validFeatures := make(map[string]bool, len(app.GetFeatures()))
	for _, f := range app.GetFeatures() {
		validFeatures[string(f)] = true
	}
	for feature := range req.Features {
		if !validFeatures[feature] {
			ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid feature: %s", feature)))
			return
		}
	}

	if err := s.features.ToggleForAllOrgs(ctx, req.Features); err != nil {
		ctx.Error(fmt.Errorf("unable to toggle feature for all orgs: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}
