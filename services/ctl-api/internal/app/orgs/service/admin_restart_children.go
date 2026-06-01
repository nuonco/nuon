package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type RestartOrgChildrenRequest struct{}

// @ID						AdminRestartOrgChildren
// @Summary				restart an org and all its children
// @Description.markdown	restart_org_children.md
// @Param					org_id	path	string						true	"org ID"
// @Param					req		body	RestartOrgChildrenRequest	true	"Input"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{boolean}	true
// @Router					/v1/orgs/{org_id}/admin-restart-children [POST]
func (s *service) RestartOrgChildren(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req RestartOrgRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	_, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
