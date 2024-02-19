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

// @ID GetAllInstalls
// @Summary	get all installs for all orgs
// @Description.markdown	get_all_installs.md
// @Tags			installs/admin
// @Accept			json
// @Param   limit  query int	 false	"limit of installs to return"	     Default(60)
// @Produce		json
// @Success		200	{array}	app.Install
// @Router			/v1/installs [get]
func (s *service) GetAllInstalls(ctx *gin.Context) {
	limitStr := ctx.DefaultQuery("limit", "60")
	limitVal, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}

	installs, err := s.getAllInstalls(ctx, limitVal)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAllInstalls(ctx context.Context, limitVal int) ([]*app.Install, error) {
	var installs []*app.Install
	res := s.db.WithContext(ctx).
		Preload("AppSandboxConfig").
		Preload("CreatedBy").
		Preload("AWSAccount").
		Preload("App").
		Preload("App.Org").
		Preload("App.AppSandboxConfigs").
		Joins("JOIN apps ON apps.id=installs.app_id").
		Joins("JOIN orgs ON orgs.id=apps.org_id").
		Where("sandbox_mode = ?", false).
		Order("created_at desc").
		Limit(limitVal).
		Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all installs: %w", res.Error)
	}

	return installs, nil
}
