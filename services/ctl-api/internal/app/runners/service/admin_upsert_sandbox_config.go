package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm/clause"
)

type AdminUpsertSandboxConfigRequest struct {
	JobType             string        `json:"job_type"`
	Operation           string        `json:"operation"`
	Enabled             *bool         `json:"enabled"`
	Duration            time.Duration `json:"duration" swaggertype:"primitive,integer"`
	SleepDuration       time.Duration `json:"sleep_duration" swaggertype:"primitive,integer"`
	ShouldError         bool          `json:"should_error"`
	ErrorMessage        string        `json:"error_message"`
	Panic               bool          `json:"panic"`
	TriggerShutdown     bool          `json:"trigger_shutdown"`
	LogTemplate         string        `json:"log_template"`
	PlanTemplate        string        `json:"plan_template"`
	PlanDisplayTemplate string        `json:"plan_display_template"`
	StateTemplate       string        `json:"state_template"`
	OutputTemplate      string        `json:"output_template"`
}

func (s *service) AdminUpsertSandboxConfig(ctx *gin.Context) {
	var req AdminUpsertSandboxConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	jobType := ctx.Param("job_type")
	if jobType == "" {
		jobType = req.JobType
	}
	if jobType == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("job_type is required")))
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	config := app.SandboxModeJobConfig{
		JobType:             jobType,
		Operation:           req.Operation,
		Enabled:             enabled,
		Duration:            req.Duration,
		SleepDuration:       req.SleepDuration,
		ShouldError:         req.ShouldError,
		ErrorMessage:        req.ErrorMessage,
		Panic:               req.Panic,
		TriggerShutdown:     req.TriggerShutdown,
		LogTemplate:         req.LogTemplate,
		PlanTemplate:        req.PlanTemplate,
		PlanDisplayTemplate: req.PlanDisplayTemplate,
		StateTemplate:       req.StateTemplate,
		OutputTemplate:      req.OutputTemplate,
	}

	if res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "job_type"}, {Name: "operation"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled", "duration", "sleep_duration", "should_error", "error_message", "panic", "trigger_shutdown", "log_template", "plan_template", "plan_display_template", "state_template", "output_template", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to upsert sandbox config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, config)
}
