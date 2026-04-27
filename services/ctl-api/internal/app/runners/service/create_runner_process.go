package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type CreateRunnerProcessRequest struct {
	Type    app.RunnerProcess `json:"type" validate:"required" swaggertype:"string"`
	Version string            `json:"version"`
}

// CreateRunnerProcessResponse mirrors the newer runner-process JSON shape that
// runners on main expect. On this hotfix branch, runner processes are not
// persisted — we return a dummy payload so newer runners can deserialize
// successfully and continue to operate.
type CreateRunnerProcessResponse struct {
	ID              string              `json:"id"`
	RunnerID        string              `json:"runner_id"`
	Type            app.RunnerProcess   `json:"type,omitempty" swaggertype:"string"`
	Version         string              `json:"version,omitempty"`
	LogStreamID     string              `json:"log_stream_id,omitempty"`
	CompositeStatus app.CompositeStatus `json:"composite_status"`
}

// @ID			CreateRunnerProcess
// @Summary	create a runner process (hotfix dummy stub — does not persist)
// @Param		req			body	CreateRunnerProcessRequest	true	"Input"
// @Param		runner_id	path	string						true	"runner ID"
// @Tags		runners/runner
// @Accept		json
// @Produce	json
// @Security	APIKey
// @Security	OrgID
// @Failure	400	{object}	stderr.ErrResponse
// @Failure	401	{object}	stderr.ErrResponse
// @Failure	403	{object}	stderr.ErrResponse
// @Failure	404	{object}	stderr.ErrResponse
// @Failure	500	{object}	stderr.ErrResponse
// @Success	201	{object}	CreateRunnerProcessResponse
// @Router		/v1/runners/{runner_id}/processes [POST]
func (s *service) CreateRunnerProcess(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req CreateRunnerProcessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	// create a log stream owned by the runner so the runner has somewhere to
	// write process logs. the runner process itself is not persisted.
	ls := app.LogStream{
		OwnerID:   runnerID,
		OwnerType: "runners",
		Open:      true,
	}
	if res := s.db.WithContext(ctx).Create(&ls); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to create log stream: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, CreateRunnerProcessResponse{
		ID:              domains.NewRunnerProcessID(),
		RunnerID:        runnerID,
		Type:            req.Type,
		Version:         req.Version,
		LogStreamID:     ls.ID,
		CompositeStatus: app.CompositeStatus{},
	})
}
