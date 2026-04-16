package service

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminUpsertSandboxSignalConfigRequest struct {
	Enabled       *bool         `json:"enabled"`
	DeadlockSleep time.Duration `json:"deadlock_sleep"`
	WorkflowSleep time.Duration `json:"workflow_sleep"`
	Panic         bool          `json:"panic"`
	Error         string        `json:"error"`
}

func (s *service) AdminUpsertSandboxSignalConfig(ctx *gin.Context) {
	signalType := ctx.Param("signal_type")
	if signalType == "" {
		ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("signal_type is required")))
		return
	}

	var req AdminUpsertSandboxSignalConfigRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	enabled := false
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	config := app.SandboxModeSignalConfig{
		SignalType:    signalType,
		Enabled:       enabled,
		DeadlockSleep: req.DeadlockSleep,
		WorkflowSleep: req.WorkflowSleep,
		Panic:         req.Panic,
		Error:         req.Error,
	}

	if res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "signal_type"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled", "deadlock_sleep", "workflow_sleep", "panic", "error", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to upsert sandbox signal config: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, config)
}
