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
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
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
// @Failure				409	{object}	stderr.ErrResponse
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

	runner, err := s.heartbeatGetRunner(ctx, runnerID)
	if err != nil {
		ctx.Error(errors.Wrap(err, "unable to get runner"))
		return
	}

	var installID, installName string
	var ownerLabels map[string]string
	switch runner.RunnerGroup.OwnerType {
	case plugins.TableName(s.db, app.Install{}):
		installID = runner.RunnerGroup.OwnerID
		if install := s.heartbeatGetInstall(ctx, installID); install != nil {
			installName = install.Name
			ownerLabels = install.Labels
		}
	case plugins.TableName(s.db, app.Org{}):
		ownerLabels = runner.Org.Labels
	}

	tagMap := make(map[string]string, len(ownerLabels)+8)
	for k, v := range ownerLabels {
		tagMap[k] = v
	}
	tagMap["org_id"] = runner.OrgID
	tagMap["org_name"] = runner.Org.Name
	tagMap["runner_id"] = runnerID
	tagMap["runner_type"] = string(runner.RunnerGroup.Type)
	tagMap["runner_version"] = req.Version
	tagMap["process_type"] = string(req.Process)
	tagMap["install_id"] = installID
	tagMap["install_name"] = installName
	tags := metrics.ToTags(tagMap)

	s.mw.Incr("runner.heart_beat", tags)
	s.mw.Timing("runner.heart_beat.alive_time", req.AliveTime, tags)

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
	if cached, ok := s.runnerHeartbeatCache.Runners.Get(runnerID); ok {
		return cached, nil
	}

	// NOTE(fd): same as getRunner w/out the RunnerGroup.Settings preload. this is hit often enough we care to optimize.
	runner := app.Runner{}
	res := s.db.WithContext(ctx).
		Preload("Org").
		Preload("RunnerGroup").
		First(&runner, "id = ?", runnerID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}

	s.runnerHeartbeatCache.Runners.Add(runnerID, &runner)
	return &runner, nil
}

func (s *service) heartbeatGetInstall(ctx context.Context, installID string) *app.Install {
	if cached, ok := s.runnerHeartbeatCache.Installs.Get(installID); ok {
		return cached
	}

	var install app.Install
	if err := s.db.WithContext(ctx).
		Select("id", "name", "labels").
		First(&install, "id = ?", installID).Error; err != nil {
		s.l.Warn("unable to look up install for heartbeat metric tag",
			zap.String("install_id", installID),
			zap.Error(err),
		)
		return nil
	}

	s.runnerHeartbeatCache.Installs.Add(installID, &install)
	return &install
}
