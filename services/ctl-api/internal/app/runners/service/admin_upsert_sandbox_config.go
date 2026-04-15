package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm/clause"
)

type AdminUpsertSandboxConfigRequest struct {
	JobType         string        `json:"job_type" validate:"required"`
	Preset          string        `json:"preset"`
	Duration        time.Duration `json:"duration"`
	FaultRate       float64       `json:"fault_rate"`
	ErrorMessage    string        `json:"error_message"`
	FailAtStep      string        `json:"fail_at_step"`
	SleepDuration   time.Duration `json:"sleep_duration"`
	Timeout         time.Duration `json:"timeout"`
	TriggerShutdown bool          `json:"trigger_shutdown"`
	LogLines        []string      `json:"log_lines"`
	PlanContents    string        `json:"plan_contents"`
	Outputs         any           `json:"outputs"`
}

func (s *service) AdminUpsertSandboxConfig(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminUpsertSandboxConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if req.JobType == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("job_type is required")))
		return
	}

	// Serialize log lines to JSON
	var logLinesJSON []byte
	if len(req.LogLines) > 0 {
		var err error
		logLinesJSON, err = json.Marshal(req.LogLines)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to marshal log lines: %w", err))
			return
		}
	}

	// Serialize outputs to JSON
	var outputsJSON []byte
	if req.Outputs != nil {
		var err error
		outputsJSON, err = json.Marshal(req.Outputs)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to marshal outputs: %w", err))
			return
		}
	}

	config := app.SandboxConfig{
		RunnerID:        runnerID,
		JobType:         req.JobType,
		Preset:          req.Preset,
		Duration:        req.Duration,
		FaultRate:       req.FaultRate,
		ErrorMessage:    req.ErrorMessage,
		FailAtStep:      req.FailAtStep,
		SleepDuration:   req.SleepDuration,
		Timeout:         req.Timeout,
		TriggerShutdown: req.TriggerShutdown,
		LogLines:        logLinesJSON,
		PlanContents:    req.PlanContents,
		Outputs:         outputsJSON,
	}

	if res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "runner_id"}, {Name: "job_type"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"preset", "duration", "fault_rate", "error_message", "fail_at_step", "sleep_duration", "timeout", "trigger_shutdown", "log_lines", "plan_contents", "outputs", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to upsert sandbox config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, config)
}
