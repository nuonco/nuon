package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ReprovisionInstallRequest struct{}

func (c *ReprovisionInstallRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID ReprovisionInstall
// @Summary	reprovision an install
// @Description.markdown	reprovision_install.md
// @Param			install_id	path	string					true	"install ID"
// @Param			req		body	ReprovisionInstallRequest	true	"Input"
// @Tags			installs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201	{string}	ok
// @Router			/v1/installs/{install_id}/reprovision [post]
func (s *service) ReprovisionInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	_, err := s.getInstall(ctx, installID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Reprovision(ctx, installID)
	ctx.JSON(http.StatusCreated, "ok")
}
