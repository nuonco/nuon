package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
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

	var runnerJobConfigs []app.SandboxModeConfig
	var signalConfigs []app.SandboxModeSignalConfig
	var stackConfig *app.SandboxModeConfig

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

	s.l.Info("upsert signal config",
		zap.String("signal_type", signalType),
		zap.String("content_type", c.ContentType()),
		zap.String("enabled", c.PostForm("enabled")),
		zap.String("deadlock", c.PostForm("deadlock_sleep_seconds")),
		zap.String("workflow", c.PostForm("workflow_sleep_seconds")),
		zap.String("panic", c.PostForm("panic")),
		zap.String("error_val", c.PostForm("error")),
	)

	deadlockSec, _ := strconv.ParseFloat(c.PostForm("deadlock_sleep_seconds"), 64)
	workflowSec, _ := strconv.ParseFloat(c.PostForm("workflow_sleep_seconds"), 64)

	config := app.SandboxModeSignalConfig{
		SignalType:    signalType,
		Enabled:       c.PostForm("enabled") == "on" || c.PostForm("enabled") == "true",
		DeadlockSleep: time.Duration(deadlockSec * float64(time.Second)),
		WorkflowSleep: time.Duration(workflowSec * float64(time.Second)),
		Panic:         c.PostForm("panic") == "on" || c.PostForm("panic") == "true",
		Error:         c.PostForm("error"),
	}

	if res := s.db.WithContext(c.Request.Context()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "signal_type"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled", "deadlock_sleep", "workflow_sleep", "panic", "error", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		s.l.Error("failed to upsert signal config", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}

	s.l.Info("signal config saved", zap.String("id", config.ID), zap.String("signal_type", signalType))

	// Re-read the saved config to get the full record with ID/timestamps
	var saved app.SandboxModeSignalConfig
	s.db.WithContext(c.Request.Context()).
		Where(app.SandboxModeSignalConfig{SignalType: signalType}).
		First(&saved)

	// Return re-rendered HTML row (open state so user sees feedback)
	component := views.SignalRowSaved(signalType, &saved)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) SandboxModeUpsertRunnerJobConfig(c *gin.Context) {
	jobType := c.Param("job_type")

	s.l.Info("upsert runner job config",
		zap.String("job_type", jobType),
		zap.String("content_type", c.ContentType()),
		zap.String("enabled", c.PostForm("enabled")),
		zap.String("duration_ms", c.PostForm("duration_ms")),
	)

	durationMs, _ := strconv.ParseInt(c.PostForm("duration_ms"), 10, 64)
	sleepMs, _ := strconv.ParseInt(c.PostForm("sleep_duration_ms"), 10, 64)
	timeoutMs, _ := strconv.ParseInt(c.PostForm("timeout_ms"), 10, 64)
	faultPct, _ := strconv.ParseFloat(c.PostForm("fault_rate_pct"), 64)

	config := app.SandboxModeConfig{
		RunnerID:        "",
		JobType:         jobType,
		Enabled:         c.PostForm("enabled") == "on",
		Preset:          c.PostForm("preset"),
		Duration:        time.Duration(durationMs) * time.Millisecond,
		FaultRate:       faultPct / 100,
		ErrorMessage:    c.PostForm("error_message"),
		FailAtStep:      c.PostForm("fail_at_step"),
		SleepDuration:   time.Duration(sleepMs) * time.Millisecond,
		Timeout:         time.Duration(timeoutMs) * time.Millisecond,
		TriggerShutdown: c.PostForm("trigger_shutdown") == "on",
	}

	if res := s.db.WithContext(c.Request.Context()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "runner_id"}, {Name: "job_type"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled", "preset", "duration", "fault_rate", "error_message", "fail_at_step", "sleep_duration", "timeout", "trigger_shutdown", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		s.l.Error("failed to upsert runner job config", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}

	// Re-read the saved config to get the full record
	var saved app.SandboxModeConfig
	s.db.WithContext(c.Request.Context()).
		Where(app.SandboxModeConfig{RunnerID: "", JobType: jobType}).
		First(&saved)

	// Return re-rendered row (open state so user sees feedback)
	templates := sandboxmode.DefaultSandboxTemplates()
	component := views.RunnerJobRowSaved(jobType, &saved, templates.LogTemplates, templates.PlanTemplates)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) SandboxModeDisableAllSignals(c *gin.Context) {
	if res := s.db.WithContext(c.Request.Context()).
		Model(&app.SandboxModeSignalConfig{}).
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
		Model(&app.SandboxModeConfig{}).
		Where("enabled = ?", true).
		Update("enabled", false); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	c.Header("HX-Trigger", "runnerJobConfigUpdated")
	c.JSON(http.StatusOK, app.EmptyResponse{})
}

func (s *service) getSandboxRunnerJobConfigs(ctx context.Context) ([]app.SandboxModeConfig, error) {
	var configs []app.SandboxModeConfig
	if res := s.db.WithContext(ctx).Order("job_type asc").Find(&configs); res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox configs: %w", res.Error)
	}
	return configs, nil
}

func (s *service) getSandboxSignalConfigs(ctx context.Context) ([]app.SandboxModeSignalConfig, error) {
	var configs []app.SandboxModeSignalConfig
	if res := s.db.WithContext(ctx).Order("signal_type asc").Find(&configs); res.Error != nil {
		return nil, fmt.Errorf("unable to get sandbox signal configs: %w", res.Error)
	}
	return configs, nil
}

func (s *service) getSandboxStackConfig(ctx context.Context) (*app.SandboxModeConfig, error) {
	var cfg app.SandboxModeConfig
	if res := s.db.WithContext(ctx).
		Where(app.SandboxModeConfig{JobType: "sandbox-terraform"}).
		First(&cfg); res.Error != nil {
		return nil, res.Error
	}
	return &cfg, nil
}
