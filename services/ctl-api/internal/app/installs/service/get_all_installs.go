package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAllInstalls
// @Summary				get all installs for all orgs
// @Description.markdown	get_all_installs.md
// @Tags					installs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					type						query	string	false	"type of installs to return"	Default(real)
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Produce				json
// @Success				200	{array}	app.Install
// @Router					/v1/installs [get]
func (s *service) GetAllInstalls(ctx *gin.Context) {
	// TODO: remove after pagination is enabled
	limitStr := ctx.DefaultQuery("limit", "60")
	limitVal, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("invalid limit %s: %w", limitStr, err),
			Description: "invalid limit",
		})
		return
	}
	orgTyp := ctx.DefaultQuery("type", "real")

	installs, err := s.getAllInstalls(ctx, limitVal, orgTyp)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get installs for: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAllInstalls(ctx *gin.Context, limitVal int, orgTyp string) ([]*app.Install, error) {
	var installs []*app.Install
	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("AppSandboxConfig").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("App").
		Preload("App.Org").
		Preload("App.AppSandboxConfigs").
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Joins("JOIN apps ON apps.id="+views.TableOrViewName(s.db, &app.Install{}, ".app_id")).
		Joins("JOIN orgs ON orgs.id=apps.org_id").
		Where("org_type = ?", orgTyp).
		Order("created_at desc").
		Limit(limitVal).
		Find(&installs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all installs: %w", res.Error)
	}

	installs, err := db.HandlePaginatedResponse(ctx, installs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installs, nil
}
