package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID GetCloudPlatformRegions
// @Summary	Get regions for a cloud platform
// @Description.markdown	get_cloud_platform_regions.md
// @Tags			general
// @Accept			json
// @Param			cloud_platform path	string	true	"cloud platform"
// @Produce		json
// @Failure		400	{object}	stderr.ErrResponse
// @Failure		401	{object}	stderr.ErrResponse
// @Failure		403	{object}	stderr.ErrResponse
// @Failure		404	{object}	stderr.ErrResponse
// @Failure		500	{object}	stderr.ErrResponse
// @Success		200	{array}	app.CloudPlatformRegion
// @Router			/v1/general/cloud-platform/{cloud_platform}/regions [GET]
func (s *service) GetCloudPlatformRegions(ctx *gin.Context) {
	platform := ctx.Param("cloud_platform")
	cloudPlatform, err := app.NewCloudPlatform(platform)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         err,
			Description: "invalid cloud platform",
		})
		return
	}

	ctx.JSON(http.StatusOK, cloudPlatform.Regions())
}
