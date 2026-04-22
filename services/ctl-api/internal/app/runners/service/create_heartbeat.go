package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/processjobsignals"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type CreateRunnerHeartBeatRequest struct {
	AliveTime time.Duration `json:"alive_time" validate:"required" swaggertype:"primitive,integer"`
	// Making this required might break existing installs? Should update all installs to send this, then make it required?
	Version   string                `json:"version"`
	Process   app.RunnerProcessType `json:"process" swaggertype:"string"`
	ProcessID string                `json:"process_id"`
}

// @ID						CreateRunnerHeartBeat
// @Summary				create a runner heart beat
// @Description.markdown	create_runner_heart_beat.md
// @Param					req			body	CreateRunnerHeartBeatRequest	true	"Input"
// @Param					runner_id	path	string							true	"runner job ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.RunnerHeartBeat
// @Router					/v1/runners/{runner_id}/heart-beats [POST]
func (s *service) CreateRunnerHeartBeat(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req CreateRunnerHeartBeatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	heartBeat, err := s.createRunnerHeartBeat(ctx, runnerID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner heart beat: %w", err))
		return
	}

	// Trigger initial health check on first heartbeat for this process
	if req.ProcessID != "" {
		if err := s.helpers.MaybeEnqueueInitialHealthCheck(ctx, runnerID, req.ProcessID); err != nil {
			s.l.Warn("unable to maybe enqueue initial health check", zap.String("process_id", req.ProcessID), zap.Error(err))
		}
	}

	// If AliveTime is short, the runner recently restarted. Wake any in-progress
	// ProcessJob workflows so they detect the restart and cancel in-flight
	// executions. The threshold mirrors the maxAliveTime check in
	// monitorJobExecution (1 minute).
	if req.AliveTime < time.Minute {
		s.helpers.SignalProcessJobsForRunner(ctx, s.l, runnerID, processjobsignals.ReasonRunnerRestarted)
	}

	runner, err := s.heartbeatGetRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner"))
		return
	}

	tags := metrics.ToTags(map[string]string{
		"org_name":       runner.Org.Name,
		"install_type":   string(runner.RunnerGroup.Type),
		"runner_version": req.Version,
		"process":        string(req.Process),
	})

	s.mw.Incr("heart_beat.incr", tags)
	s.mw.Timing("heart_beat.alive_time", req.AliveTime, tags)

	ctx.JSON(http.StatusCreated, heartBeat)
}

func (s *service) createRunnerHeartBeat(ctx context.Context, runnerID string, req CreateRunnerHeartBeatRequest) (*app.RunnerHeartBeat, error) {
	runnerHeartBeat := app.RunnerHeartBeat{
		RunnerID:  runnerID,
		ProcessID: req.ProcessID,
		AliveTime: req.AliveTime,
		Version:   req.Version,
	}
	// if we do not receive a value, set a default
	if req.Process != "" {
		runnerHeartBeat.Process = req.Process
	} else {
		runnerHeartBeat.Process = app.RunnerProcessTypeUnknown
	}

	res := s.chDB.
		WithContext(ctx).
		Create(&runnerHeartBeat)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create runner heart beat: %w", res.Error)
	}

	return &runnerHeartBeat, nil
}

func (s *service) heartbeatGetRunner(ctx context.Context, runnerID string) (*app.Runner, error) {
	// NOTE(fd): same as getRunner w/out the RunnerGroup.Settings preload. this is hit often enough we care to optimize.
	runner := app.Runner{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("RunnerGroup").
		First(&runner, "id = ?", runnerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	return &runner, nil
}
