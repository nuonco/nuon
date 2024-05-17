package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID GetAllInstallers
// @Summary	get all installers for all orgs
// @Description.markdown	get_all_installers.md
// @Tags			installers/admin
// @Accept			json
// @Param   limit  query int	 false	"limit of installers to return"	     Default(60)
// @Produce		json
// @Success		200	{array}	app.AppInstaller
// @Router			/v1/installers [get]
func (s *service) GetAllInstallers(ctx *gin.Context) {
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

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAllInstallers(ctx context.Context, limitVal int) ([]*app.Installer, error) {
	var installers []*app.Installer
	res := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Metadata").
		Preload("Apps").
		Preload("Apps.Org").
		Order("created_at desc").
		Limit(limitVal).
		Find(&installers)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get installers: %w", res.Error)
	}

	return installers, nil
}
