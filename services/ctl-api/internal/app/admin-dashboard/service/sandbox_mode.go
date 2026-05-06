package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/sandboxmode"
	sbtemplates "github.com/nuonco/nuon/services/ctl-api/internal/pkg/sandboxmode/templates"
)

func (s *service) SandboxMode(c *gin.Context) {
	ctx := c.Request.Context()

	runnerJobConfigs, err := s.getSandboxRunnerJobConfigs(ctx)
	if err != nil {
		s.l.Warn("failed to get runner job configs", zap.Error(err))
	}
	if runnerJobConfigs == nil {
		runnerJobConfigs = []app.SandboxModeJobConfig{}
	}

	signalConfigs, err := s.getSandboxSignalConfigs(ctx)
	if err != nil {
		s.l.Warn("failed to get signal configs", zap.Error(err))
	}
	if signalConfigs == nil {
		signalConfigs = []app.SandboxModeSignalConfig{}
	}

	stackConfig, _ := s.getSandboxStackConfig(ctx)

	c.JSON(http.StatusOK, gin.H{
		"runner_job_configs":             runnerJobConfigs,
		"signal_configs":                 signalConfigs,
		"stack_config":                   stackConfig,
		"all_signal_types":               signals.AllSignalTypes(),
		"all_runner_job_types":           sandboxmode.AllRunnerJobTypes(),
		"all_runner_job_operation_types": sandboxmode.AllRunnerJobOperationTypes(),
		"templates":                      sbtemplates.AllTemplates(),
		"flow_templates":                 sbtemplates.FlowTemplates(),
	})
}

func (s *service) SandboxModeRunnerJobsTable(c *gin.Context) {
	ctx := c.Request.Context()
	configs, err := s.getSandboxRunnerJobConfigs(ctx)
	if err != nil {
		s.l.Error("failed to get runner job configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configs":              configs,
		"all_runner_job_types": sandboxmode.AllRunnerJobTypes(),
		"templates":            sbtemplates.AllTemplates(),
	})
}

func (s *service) SandboxModeRunnerJobsRows(c *gin.Context) {
	ctx := c.Request.Context()
	configs, err := s.getSandboxRunnerJobConfigs(ctx)
	if err != nil {
		s.l.Error("failed to get runner job configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configs":              configs,
		"all_runner_job_types": sandboxmode.AllRunnerJobTypes(),
	})
}

func (s *service) SandboxModeBuilder(c *gin.Context) {
	jobType := c.Query("job_type")

	var cfg *app.SandboxModeJobConfig
	if jobType != "" {
		var found app.SandboxModeJobConfig
		if res := s.db.WithContext(c.Request.Context()).
			Where(app.SandboxModeJobConfig{JobType: jobType}).
			First(&found); res.Error == nil {
			cfg = &found
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"job_type":             jobType,
		"config":               cfg,
		"templates":            sbtemplates.AllTemplates(),
		"all_runner_job_types": sandboxmode.AllRunnerJobTypes(),
	})
}

func (s *service) SandboxModeSignalsTable(c *gin.Context) {
	ctx := c.Request.Context()
	configs, err := s.getSandboxSignalConfigs(ctx)
	if err != nil {
		s.l.Error("failed to get signal configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configs":          configs,
		"all_signal_types": signals.AllSignalTypes(),
	})
}

func (s *service) SandboxModeSignalRows(c *gin.Context) {
	ctx := c.Request.Context()
	configs, err := s.getSandboxSignalConfigs(ctx)
	if err != nil {
		s.l.Error("failed to get signal configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"configs":          configs,
		"all_signal_types": signals.AllSignalTypes(),
	})
}

func (s *service) SandboxModeStacksTable(c *gin.Context) {
	ctx := c.Request.Context()
	cfg, err := s.getSandboxStackConfig(ctx)
	if err != nil {
		s.l.Error("failed to get stack config", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"config":    cfg,
		"templates": sbtemplates.AllTemplates(),
	})
}

type upsertSignalConfigRequest struct {
	Enabled              bool    `json:"enabled"`
	DeadlockSleepSeconds float64 `json:"deadlock_sleep_seconds"`
	WorkflowSleepSeconds float64 `json:"workflow_sleep_seconds"`
	Panic                bool    `json:"panic"`
	Error                string  `json:"error"`
	ValidateError        string  `json:"validate_error"`
}

func (s *service) SandboxModeUpsertSignalConfig(c *gin.Context) {
	signalType := c.Param("signal_type")

	var req upsertSignalConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := app.SandboxModeSignalConfig{
		CreatedByID:   createdByIDFromGinContext(c),
		SignalType:    signalType,
		Enabled:       req.Enabled,
		DeadlockSleep: time.Duration(req.DeadlockSleepSeconds * float64(time.Second)),
		WorkflowSleep: time.Duration(req.WorkflowSleepSeconds * float64(time.Second)),
		Panic:         req.Panic,
		Error:         req.Error,
		ValidateError: req.ValidateError,
	}

	if res := s.db.WithContext(c.Request.Context()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "signal_type"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled", "deadlock_sleep", "workflow_sleep", "panic", "error", "validate_error", "updated_at"}),
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

	c.JSON(http.StatusOK, gin.H{
		"config":      &saved,
		"signal_type": signalType,
	})
}

type upsertRunnerJobConfigRequest struct {
	Operation           string `json:"operation"`
	Enabled             bool   `json:"enabled"`
	DurationMs          int64  `json:"duration_ms"`
	SleepDurationMs     int64  `json:"sleep_duration_ms"`
	ShouldError         bool   `json:"should_error"`
	Panic               bool   `json:"panic"`
	TriggerShutdown     bool   `json:"trigger_shutdown"`
	LogTemplate         string `json:"log_template"`
	PlanTemplate        string `json:"plan_template"`
	PlanDisplayTemplate string `json:"plan_display_template"`
	StateTemplate       string `json:"state_template"`
	OutputTemplate      string `json:"output_template"`
}

func (s *service) SandboxModeUpsertRunnerJobConfig(c *gin.Context) {
	jobType := c.Param("job_type")

	var req upsertRunnerJobConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config := app.SandboxModeJobConfig{
		CreatedByID:         createdByIDFromGinContext(c),
		JobType:             jobType,
		Operation:           req.Operation,
		Enabled:             req.Enabled,
		Duration:            time.Duration(req.DurationMs) * time.Millisecond,
		SleepDuration:       time.Duration(req.SleepDurationMs) * time.Millisecond,
		ShouldError:         req.ShouldError,
		Panic:               req.Panic,
		TriggerShutdown:     req.TriggerShutdown,
		LogTemplate:         req.LogTemplate,
		PlanTemplate:        req.PlanTemplate,
		PlanDisplayTemplate: req.PlanDisplayTemplate,
		StateTemplate:       req.StateTemplate,
		OutputTemplate:      req.OutputTemplate,
	}

	if res := s.db.WithContext(c.Request.Context()).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "job_type"}, {Name: "operation"}, {Name: "deleted_at"}},
			DoUpdates: clause.AssignmentColumns([]string{"enabled", "duration", "sleep_duration", "should_error", "panic", "trigger_shutdown", "log_template", "plan_template", "plan_display_template", "state_template", "output_template", "updated_at"}),
		}).
		Create(&config); res.Error != nil {
		s.l.Error("failed to upsert runner job config", zap.Error(res.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}

	// Re-read the saved config to get the full record (match both job_type and operation)
	var saved app.SandboxModeJobConfig
	s.db.WithContext(c.Request.Context()).
		Where(map[string]interface{}{"job_type": jobType, "operation": req.Operation}).
		First(&saved)

	c.JSON(http.StatusOK, gin.H{
		"config":   &saved,
		"job_type": jobType,
	})
}

func (s *service) SandboxModeDisableAllSignals(c *gin.Context) {
	if res := s.db.WithContext(c.Request.Context()).
		Model(&app.SandboxModeSignalConfig{}).
		Where("enabled = ?", true).
		Update("enabled", false); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "disabled"})
}

func (s *service) SandboxModeDeleteRunnerJobConfig(c *gin.Context) {
	configID := c.Param("config_id")
	if res := s.db.WithContext(c.Request.Context()).
		Where("id = ?", configID).
		Delete(&app.SandboxModeJobConfig{}); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (s *service) SandboxModeDisableAllRunnerJobs(c *gin.Context) {
	if res := s.db.WithContext(c.Request.Context()).
		Model(&app.SandboxModeJobConfig{}).
		Where("enabled = ?", true).
		Update("enabled", false); res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "disabled"})
}

func (s *service) SandboxModeApplyFlowTemplate(c *gin.Context) {
	templateKey := c.Param("template_key")

	flow := sbtemplates.FindFlowTemplate(templateKey)
	if flow == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found: " + templateKey})
		return
	}

	createdBy := createdByIDFromGinContext(c)

	for _, fc := range flow.Configs {
		config := app.SandboxModeJobConfig{
			CreatedByID:         createdBy,
			JobType:             fc.JobType,
			Enabled:             fc.Enabled,
			Duration:            time.Duration(fc.DurationMs) * time.Millisecond,
			LogTemplate:         fc.LogTemplate,
			PlanTemplate:        fc.PlanTemplate,
			PlanDisplayTemplate: fc.PlanDisplayTemplate,
			StateTemplate:       fc.StateTemplate,
			OutputTemplate:      fc.OutputTemplate,
		}

		if res := s.db.WithContext(c.Request.Context()).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "job_type"}, {Name: "operation"}, {Name: "deleted_at"}},
				DoUpdates: clause.AssignmentColumns([]string{"enabled", "duration", "log_template", "plan_template", "plan_display_template", "state_template", "output_template", "updated_at"}),
			}).
			Create(&config); res.Error != nil {
			s.l.Error("failed to apply flow template config", zap.String("job_type", fc.JobType), zap.Error(res.Error))
			c.JSON(http.StatusInternalServerError, gin.H{"error": res.Error.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "applied", "template": templateKey})
}

func (s *service) getSandboxRunnerJobConfigs(ctx context.Context) ([]app.SandboxModeJobConfig, error) {
	var configs []app.SandboxModeJobConfig
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

func (s *service) getSandboxStackConfig(ctx context.Context) (*app.SandboxModeJobConfig, error) {
	var cfg app.SandboxModeJobConfig
	if res := s.db.WithContext(ctx).
		Where(app.SandboxModeJobConfig{JobType: "sandbox-terraform"}).
		First(&cfg); res.Error != nil {
		return nil, res.Error
	}
	return &cfg, nil
}

func boolToFloat(b bool) float64 {
	if b {
		return 1.0
	}
	return 0
}

func createdByIDFromGinContext(c *gin.Context) string {
	if acctID, exists := c.Get("account_id"); exists {
		if s, ok := acctID.(string); ok && s != "" {
			return s
		}
	}
	return "admin-dashboard"
}
