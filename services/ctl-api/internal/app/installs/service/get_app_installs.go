package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetAppInstalls
// @Summary				get all installs for an app
// @Description.markdown	get_app_installs.md
// @Param					app_id						path	string	true	"app ID"
// @Param					q							query	string	false	"search query"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.Install
// @Router					/v1/apps/{app_id}/installs [GET]
func (s *service) GetAppInstalls(ctx *gin.Context) {
	appID := ctx.Param("app_id")
	q := ctx.Query("q")

	installs, err := s.getAppInstalls(ctx, appID, q)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, installs)
}

func (s *service) getAppInstalls(ctx *gin.Context, appID string, q string) ([]app.Install, error) {
	var installs []app.Install
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination)

	if q != "" {
		tx = tx.Where("name ILIKE ?", "%"+q+"%")
	}

	tx = tx.Where("app_id = ?", appID).
		Preload("AppSandboxConfig").
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).
		Preload("AWSAccount").
		Preload("AppRunnerConfig").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Runners").
		Order("name ASC")

	res := tx.Find(&installs)

	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	installs, err := db.HandlePaginatedResponse(ctx, installs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return installs, nil
}
