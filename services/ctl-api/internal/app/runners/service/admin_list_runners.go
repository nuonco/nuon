package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID			AdminGetAllRunners
// @BasePath	/v1/runners
// @Summary	Return all runners
// @Schemes
// @Description	return all orgs
// @Param			type						query	string	false	"type of runner to return"		Default(org)
// @Param			offset						query	int		false	"offset of results to return"	Default(0)
// @Param			limit						query	int		false	"limit of results to return"	Default(10)
// @Param			page						query	int		false	"page number of results to return"	Default(0)
// @Tags			runners/admin
// @Security		AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{array}	app.Runner
// @Router			/v1/runners [GET]
func (s *service) AdminGetAllRunners(ctx *gin.Context) {
	runnerTyp := ctx.DefaultQuery("type", "org")

	runners, err := s.getAllRunners(ctx, runnerTyp)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, runners)
}

func (s *service) getAllRunners(ctx *gin.Context, typ string) ([]*app.Runner, error) {
	var runners []*app.Runner

	res := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Preload("CreatedBy").
		Joins("JOIN runner_groups ON runner_groups.id = runners.runner_group_id").
		Where("runner_groups.type = ?", typ).
		Order("created_at desc").
		Find(&runners)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get all runners: %w", res.Error)
	}

	runners, err := db.HandlePaginatedResponse(ctx, runners)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return runners, nil
}
