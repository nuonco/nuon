package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AdminDeleteOrgRequest struct {}

// NOTE(jm): this endpoint is intentionally not documented as it is extremely destructive.
func (s *service) AdminDeleteOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	err := s.deleteOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	s.hooks.Deleted(ctx, orgID)
	ctx.JSON(http.StatusOK, true)
}
