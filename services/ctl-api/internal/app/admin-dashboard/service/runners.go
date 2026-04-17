package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) Runners(c *gin.Context) {
	ctx := c.Request.Context()

	runners, err := s.getSandboxRunnerViews(ctx)
	if err != nil {
		s.l.Error("failed to get sandbox runners", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch runners"})
		return
	}

	component := views.Runners(runners)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func (s *service) getSandboxRunnerViews(ctx context.Context) ([]views.SandboxRunnerView, error) {
	// Find all runner groups with sandbox mode enabled
	var settings []app.RunnerGroupSettings
	if res := s.db.WithContext(ctx).
		Where(app.RunnerGroupSettings{SandboxMode: true}).
		Find(&settings); res.Error != nil {
		return nil, fmt.Errorf("unable to find sandbox settings: %w", res.Error)
	}

	if len(settings) == 0 {
		return []views.SandboxRunnerView{}, nil
	}

	groupIDs := make([]string, len(settings))
	for i, setting := range settings {
		groupIDs[i] = setting.RunnerGroupID
	}

	// Find only active runners in those groups, preload RunnerGroup for install context
	var runners []app.Runner
	if res := s.db.WithContext(ctx).
		Where("runner_group_id IN ? AND status = ?", groupIDs, app.RunnerStatusActive).
		Preload("RunnerGroup").
		Find(&runners); res.Error != nil {
		return nil, fmt.Errorf("unable to find sandbox runners: %w", res.Error)
	}

	result := make([]views.SandboxRunnerView, 0, len(runners))
	for _, runner := range runners {
		view := views.SandboxRunnerView{
			Runner: runner,
		}

		// Resolve the install name if the runner group owner is an install
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
			Where("runner_id = ?", runner.ID).
			Order("created_at desc").
			First(&process); res.Error == nil {
			view.ProcessOnline = process.ProcessStatus() == app.RunnerProcessStatusActive
			if process.Version != "" {
				view.Version = process.Version
			}
		}

		// Get sandbox configs for this runner
		var configs []app.SandboxModeJobConfig
		if res := s.db.WithContext(ctx).
			Where("job_type != ''").
			Order("job_type").
			Find(&configs); res.Error == nil {
			view.Configs = configs
		}

		result = append(result, view)
	}

	return result, nil
}
