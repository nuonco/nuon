package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/powertoolsdev/mono/pkg/types/state"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID                       GetInstallStateHistory
// @Summary                  Get install state history.
// @Description.markdown     get_install_state_history.md
// @Param                    install_id                 path    string  true  "install ID"
// @Param                    offset                     query   int     false "offset of results to return"    Default(0)
// @Param                    limit                      query   int     false "limit of results to return"     Default(10)
// @Param                    page                       query   int     false "page number of results to return" Default(0)
// @Param                    x-nuon-pagination-enabled  header  bool    false "Enable pagination"
// @Tags                     installs
// @Accept                   json
// @Produce                  json
// @Security                 APIKey
// @Security                 OrgID
// @Failure                  400 {object} stderr.ErrResponse
// @Failure                  401 {object} stderr.ErrResponse
// @Failure                  403 {object} stderr.ErrResponse
// @Failure                  404 {object} stderr.ErrResponse
// @Failure                  500 {object} stderr.ErrResponse
// @Success                  200 {array} app.InstallState
// @Router                   /v1/installs/{install_id}/state-history [get]
func (s *service) GetInstallStateHistory(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	states, err := s.getInstallStateHistory(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install states: %w", err))
	}
	ctx.JSON(http.StatusOK, states)
}

func (s *service) getInstallStateHistory(ctx *gin.Context, installID string) ([]*app.InstallState, error) {
	var states []*app.InstallState
	res := s.db.
		Scopes(scopes.WithOffsetPagination).
		Where("install_id = ?", installID).
		Order("created_at desc").Find(&states)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install states: %w", res.Error)
	}

	states, err := db.HandlePaginatedResponse(ctx, states)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return states, nil
}
