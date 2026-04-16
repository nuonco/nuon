package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/sandboxmode"
)

func (s *service) SandboxMode(c *gin.Context) {
	ctx := c.Request.Context()
	tab := c.Query("tab")
	if tab == "" {
		tab = "runner-jobs"
	}

	var runnerJobConfigs []app.SandboxConfig
	var signalConfigs []app.SandboxSignalConfig
	var stackConfig *app.SandboxConfig

	switch tab {
	case "runner-jobs":
		runnerJobConfigs, _ = s.getSandboxRunnerJobConfigs(ctx)
	case "signals":
		signalConfigs, _ = s.getSandboxSignalConfigs(ctx)
	case "stacks":
		stackConfig, _ = s.getSandboxStackConfig(ctx)
	}

	templates := sandboxmode.DefaultSandboxTemplates()
	component := views.SandboxMode(views.SandboxModeData{
		ActiveTab:         tab,
		RunnerJobConfigs:  runnerJobConfigs,
		SignalConfigs:     signalConfigs,
		StackConfig:       stackConfig,
		AllSignalTypes:    signals.AllSignalTypes(),
		AllRunnerJobTypes: sandboxmode.AllRunnerJobTypes(),
		LogTemplates:      templates.LogTemplates,
		PlanTemplates:     templates.PlanTemplates,
	})
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) SandboxModeRunnerJobsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	configs, err := s.getSandboxRunnerJobConfigs(ctx)
	if err != nil {
		s.l.Error("failed to get runner job configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	templates := sandboxmode.DefaultSandboxTemplates()
	component := views.SandboxModeRunnerJobsTable(configs, sandboxmode.AllRunnerJobTypes(), search, templates.LogTemplates, templates.PlanTemplates)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) SandboxModeSignalsTable(c *gin.Context) {
	ctx := c.Request.Context()
	search := c.Query("search")
	configs, err := s.getSandboxSignalConfigs(ctx)
	if err != nil {
		s.l.Error("failed to get signal configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	component := views.SandboxModeSignalsTable(configs, signals.AllSignalTypes(), search)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) SandboxModeStacksTable(c *gin.Context) {
	ctx := c.Request.Context()
	cfg, err := s.getSandboxStackConfig(ctx)
	if err != nil {
		s.l.Error("failed to get stack config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	templates := sandboxmode.DefaultSandboxTemplates()
	component := views.SandboxModeStacksTable(cfg, templates.LogTemplates, templates.PlanTemplates)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) SandboxModeUpsertSignalConfig(c *gin.Context) {
	signalType := c.Param("signal_type")

	var req struct {
		Enabled       bool   `json:"enabled"`
		DeadlockSleep int64  `json:"deadlock_sleep_seconds"`
		WorkflowSleep int64  `json:"workflow_sleep_seconds"`
		Panic         bool   `json:"panic"`
		Error         string `json:"error"`
	}
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := app.SandboxSignalConfig{
		SignalType:    signalType,
		Enabled:       req.Enabled,
		DeadlockSleep: time.Duration(req.DeadlockSleep) * time.Second,
		WorkflowSleep: time.Duration(req.WorkflowSleep) * time.Second,
		Panic:         req.Panic,
		Error:         req.Error,
	}

	if res := s.db.WithContext(c.Request.Context()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "signal_type"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled", "deadlock_sleep", "workflow_sleep", "panic", "error", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}

	c.Header("HX-Trigger", "signalConfigUpdated")
	c.JSON(http.StatusOK, config)
}

func (s *service) SandboxModeDisableAllSignals(c *gin.Context) {
	if res := s.db.WithContext(c.Request.Context()).
		Model(&app.SandboxSignalConfig{}).
		Where("enabled = ?", true).
		Update("enabled", false); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	c.Header("HX-Trigger", "signalConfigUpdated")
	c.JSON(http.StatusOK, app.EmptyResponse{})
}

func (s *service) SandboxModeDisableAllRunnerJobs(c *gin.Context) {
	if res := s.db.WithContext(c.Request.Context()).
		Model(&app.SandboxConfig{}).
		Where("enabled = ?", true).
		Update("enabled", false); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	c.Header("HX-Trigger", "runnerJobConfigUpdated")
	c.JSON(http.StatusOK, app.EmptyResponse{})
}

func (s *service) getSandboxRunnerJobConfigs(ctx context.Context) ([]app.SandboxConfig, error) {
	var configs []app.SandboxConfig
	if res := s.db.WithContext(ctx).Order("job_type asc").Find(&configs); res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox configs: %w", res.Error)
	}
	return configs, nil
}

func (s *service) getSandboxSignalConfigs(ctx context.Context) ([]app.SandboxSignalConfig, error) {
	var configs []app.SandboxSignalConfig
	if res := s.db.WithContext(ctx).Order("signal_type asc").Find(&configs); res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox signal configs: %w", res.Error)
	}
	return configs, nil
}

func (s *service) getSandboxStackConfig(ctx context.Context) (*app.SandboxConfig, error) {
	var cfg app.SandboxConfig
	if res := s.db.WithContext(ctx).
		Where(app.SandboxConfig{JobType: "sandbox-terraform"}).
		First(&cfg); res.Error != nil {
		return nil, res.Error
	}
	return &cfg, nil
}
