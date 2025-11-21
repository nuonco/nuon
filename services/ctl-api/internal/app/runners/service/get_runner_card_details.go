package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type RunnerCardDetailsResponse struct {
	Runner          app.Runner          `json:"runner"`
	LatestHeartBeat app.RunnerHeartBeat `json:"latest_heart_beat"`
}

// @ID						GetRunnerCardDetails
// @Summary				get runner card details
// @Description.markdown	get_runner_settings.md
// @Param					runner_id	path	string	true	"runner ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	RunnerCardDetailsResponse
// @Router					/v1/runners/{runner_id}/card-details [get]
func (s *service) GetRunnerCardDetails(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	runnerID := ctx.Param("runner_id")

	runner, err := s.getOrgRunner(ctx, runnerID, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	lastHeartBeat, err := s.getRunnerLatestHeartBeat(ctx, runnerID)
	if err != nil {
		ctx.Error(err)
		return
	}

	response := RunnerCardDetailsResponse{
		Runner:          *runner,
		LatestHeartBeat: *lastHeartBeat,
	}

	ctx.JSON(http.StatusOK, response)
}
