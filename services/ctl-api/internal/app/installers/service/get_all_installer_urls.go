package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID GetAllInstallersURLs
// @Summary	get all installer urls for all orgs
// @Description.markdown	get_all_installer_urls.md
// @Tags			installers/admin
// @Accept			json
// @Param   limit  query int	 false	"limit of installers to return"	     Default(60)
// @Produce		json
// @Success		200	{array}	string
// @Router			/v1/installers/urls [get]
func (s *service) GetAllInstallerURLs(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "60")
	limitVal, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}

	installs, err := s.getAllInstallers(ctx, limitVal)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for: %w", err))
		return
	}
	urls := make([]string, 0)
	for _, installer := range installs {
		urls = append(urls, installer.InstallerURL)
	}

	ctx.JSON(http.StatusOK, urls)
}
