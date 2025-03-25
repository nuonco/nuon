package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

//	@ID						AdminGetOrgFeatures
//	@Summary				get available org features
//	@Description.markdown	admin_get_org_features.md
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				200	{array}	string
//	@Router					/v1/orgs/admin-features  [GET]
func (s *service) AdminGetOrgFeatures(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, app.GetFeatures())
}
