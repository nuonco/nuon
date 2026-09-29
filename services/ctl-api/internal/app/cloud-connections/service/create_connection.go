package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CreateRequest struct {
	Name          string                    `json:"name"`
	Platform      app.CloudPlatform         `json:"platform" swaggertype:"string" enums:"aws"`
	TargetID      string                    `json:"target_id"`
	Principal     string                    `json:"principal"`
	DefaultRegion string                    `json:"default_region,omitempty"`
	Preset        app.CloudConnectionPreset `json:"preset"`
}

// @ID CreateCloudConnection
// @Summary create a cloud connection
// @Description Create an AWS connection using the stacks or custom preset. Custom renders trust only; attach your own permissions policy.
// @Tags cloud-connections
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param req body CreateRequest true "Input"
// @Failure 400 {object} stderr.ErrResponse
// @Failure 401 {object} stderr.ErrResponse
// @Failure 403 {object} stderr.ErrResponse
// @Failure 404 {object} stderr.ErrResponse
// @Failure 409 {object} stderr.ErrResponse
// @Failure 500 {object} stderr.ErrResponse
// @Success 201 {object} ConnectionResponse
// @Router /v1/cloud-connections [post]
func (s *service) Create(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var req CreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	connection := app.CloudConnection{OrgID: org.ID, Name: req.Name, Platform: req.Platform, TargetID: req.TargetID, Principal: req.Principal, DefaultRegion: req.DefaultRegion, Preset: req.Preset}
	if err := validateConnection(&connection); err != nil {
		ctx.Error(userError(err))
		return
	}
	if err := s.db.WithContext(ctx).Create(&connection).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to create cloud connection: %w", err))
		return
	}
	if _, err := s.helpers.EnsureConnectionQueue(ctx, &connection); err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, &connection)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, response)
}
