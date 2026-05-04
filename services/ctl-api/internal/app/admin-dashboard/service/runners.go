package service

import (
	"context"
	"fmt"
	"net/http"

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

	c.JSON(http.StatusOK, gin.H{
		"runners": runners,
	})
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

	// Sandbox job configs are global (no per-runner filter), so fetch once.
	var configs []app.SandboxModeJobConfig
	if err := s.db.WithContext(ctx).
		Where("job_type != ''").
		Order("job_type").
		Find(&configs).Error; err != nil {
		s.l.Warn("failed to load sandbox job configs", zap.Error(err))
	}

	// Batch-resolve install names for any install-owned runner groups.
	installIDs := make(map[string]struct{})
	for _, runner := range runners {
		if runner.RunnerGroup.OwnerType == "installs" && runner.RunnerGroup.OwnerID != "" {
			installIDs[runner.RunnerGroup.OwnerID] = struct{}{}
		}
	}
	installNames := make(map[string]string, len(installIDs))
	if len(installIDs) > 0 {
		ids := make([]string, 0, len(installIDs))
		for id := range installIDs {
			ids = append(ids, id)
		}
		var installs []app.Install
		if err := s.db.WithContext(ctx).
			Select("id", "name").
			Where("id IN ?", ids).
			Find(&installs).Error; err == nil {
			for _, i := range installs {
				installNames[i.ID] = i.Name
			}
		}
	}

	// Batch-resolve latest process per runner (one query, DISTINCT ON).
	type processInfo struct {
		Online  bool
		Version string
	}
	processMap := make(map[string]processInfo, len(runners))
	if len(runners) > 0 {
		runnerIDs := make([]string, len(runners))
		for i, r := range runners {
			runnerIDs[i] = r.ID
		}
		var processes []app.RunnerProcess
		if err := s.db.WithContext(ctx).
			Raw(`SELECT DISTINCT ON (runner_id) * FROM runner_processes
				 WHERE runner_id IN ? AND deleted_at = 0
				 ORDER BY runner_id, created_at DESC`, runnerIDs).
			Scan(&processes).Error; err == nil {
			for _, p := range processes {
				processMap[p.RunnerID] = processInfo{
					Online:  p.ProcessStatus() == app.RunnerProcessStatusActive,
					Version: p.Version,
				}
			}
		}
	}

	result := make([]views.SandboxRunnerView, 0, len(runners))
	for _, runner := range runners {
		view := views.SandboxRunnerView{
			Runner:  runner,
			Configs: configs,
		}

		if runner.RunnerGroup.OwnerType == "installs" {
			view.InstallID = runner.RunnerGroup.OwnerID
			view.InstallName = installNames[runner.RunnerGroup.OwnerID]
		}

		if pi, ok := processMap[runner.ID]; ok {
			view.ProcessOnline = pi.Online
			if pi.Version != "" {
				view.Version = pi.Version
			}
		}

		result = append(result, view)
	}

	return result, nil
}
