package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						GetOrgFeatures
// @Summary				get available org features
// @Description.markdown	get_org_features.md
// @Tags					orgs
// @Security				APIKey
// @Security				OrgID
// @Accept					json
// @Produce				json
// @Success				200	{array}	app.OrgFeatureInfo
// @Router					/v1/orgs/features  [GET]
func (s *service) GetOrgFeatures(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, app.GetFeaturesWithDescriptions())
}
