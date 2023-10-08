package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RestartOrgRequest struct{}

//	@BasePath	/v1/orgs
//
// Restart an org's event loop
//
//	@Summary	restart an orgs event loop
//	@Schemes
//	@Description	restart org event loop
//	@Param			org_id	path	string					true	"org ID"
//	@Param			req			body	RestartOrgRequest	true	"Input"
//	@Tags			orgs/admin
//	@Accept			json
//	@Produce		json
//	@Success		200	{boolean}	true
//	@Router			/v1/orgs/{org_id}/restart [POST]
func (s *service) RestartOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req RestartOrgRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	org, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org: %w", err))
		return
	}

	s.hooks.Restart(ctx, org.ID)
	ctx.JSON(http.StatusOK, true)
}
