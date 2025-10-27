package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/powertoolsdev/mono/pkg/types/state"
)

// type State = state.State

// @ID						GetInstallState
// @Summary				Get the current state of an install.
// @Description.markdown	get_install_state.md
// @Param					install_id	path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	state.State
// @Router					/v1/installs/{install_id}/state [get]
func (s *service) GetInstallState(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	is, err := s.helpers.GetInstallState(ctx, installID, true, true)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install state: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, is)
}
