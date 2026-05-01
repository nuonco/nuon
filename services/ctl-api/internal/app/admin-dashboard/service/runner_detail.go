package service

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) RunnerDetail(c *gin.Context) {
	ctx := c.Request.Context()
	runnerID := c.Param("id")

	var runner app.Runner
	if res := s.db.WithContext(ctx).
		Preload("RunnerGroup").
		Where("id = ?", runnerID).
		First(&runner); res.Error != nil {
		s.l.Error("failed to get runner", zap.Error(res.Error))
		c.JSON(http.StatusNotFound, gin.H{"error": "Runner not found"})
		return
	}

	view := views.RunnerDetailView{
		Runner:  runner,
		Configs: make(map[string]*app.SandboxModeJobConfig),
	}

	// Resolve install
	if runner.RunnerGroup.OwnerType == "installs" {
		var install app.Install
		if res := s.db.WithContext(ctx).
			Select("id", "name").
			Where("id = ?", runner.RunnerGroup.OwnerID).
			First(&install); res.Error == nil {
			view.InstallID = install.ID
			view.InstallName = install.Name
		}
	}

	// Get latest process for online status
	var process app.RunnerProcess
	if res := s.db.WithContext(ctx).
		Where("runner_id = ?", runnerID).
		Order("created_at desc").
		First(&process); res.Error == nil {
		view.Process = &process
		view.ProcessOnline = process.ProcessStatus() == app.RunnerProcessStatusActive
	}

	// Load sandbox configs keyed by job type
	var configs []app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Where("job_type != ''").
		Find(&configs); res.Error == nil {
		for i := range configs {
			view.Configs[configs[i].JobType] = &configs[i]
		}
	}

	c.JSON(http.StatusOK, view)
}

type dashboardUpsertConfigRequest struct {
	JobType         string        `json:"job_type"`
	Duration        time.Duration `json:"duration"`
	ShouldError     bool          `json:"should_error"`
	Panic           bool          `json:"panic"`
	TriggerShutdown bool          `json:"trigger_shutdown"`
}

func (s *service) RunnerUpsertConfig(c *gin.Context) {
	_ = c.Param("id")

	var req dashboardUpsertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.JobType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "job_type is required"})
		return
	}

	config := app.SandboxModeJobConfig{
		JobType:         req.JobType,
		Duration:        req.Duration,
		ShouldError:     req.ShouldError,
		Panic:           req.Panic,
		TriggerShutdown: req.TriggerShutdown,
	}

	if res := s.db.WithContext(c.Request.Context()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "job_type"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"duration", "should_error", "panic", "trigger_shutdown", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		s.l.Error("failed to upsert sandbox config", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *service) RunnerDeleteConfig(c *gin.Context) {
	_ = c.Param("id")
	jobType := c.Param("job_type")

	if res := s.db.WithContext(c.Request.Context()).
		Where(app.SandboxModeJobConfig{JobType: jobType}).
		Delete(&app.SandboxModeJobConfig{}); res.Error != nil {
		s.l.Error("failed to delete sandbox config", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (s *service) RunnerResetConfigs(c *gin.Context) {
	_ = c.Param("id")

	if res := s.db.WithContext(c.Request.Context()).
		Where("job_type != ''").
		Delete(&app.SandboxModeJobConfig{}); res.Error != nil {
		s.l.Error("failed to reset sandbox configs", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset configs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
