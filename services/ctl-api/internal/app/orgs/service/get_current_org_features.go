package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetCurrentOrgFeatures
// @Summary				get current org's feature flags
// @Description.markdown	get_current_org_features.md
// @Tags					orgs
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Success				200	{object}	map[string]bool
// @Router					/v1/orgs/current/features  [GET]
func (s *service) GetCurrentOrgFeatures(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org.Features)
}
