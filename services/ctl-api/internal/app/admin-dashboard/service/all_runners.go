package service

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) AllRunners(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Query("org_id")

	runners, err := s.getAllRunnerViews(ctx, orgID)
	if err != nil {
		s.l.Error("failed to get all runners", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch runners"})
		return
	}

	orgs := s.getOrgOptions(ctx)
	if orgs == nil {
		orgs = []views.OrgOption{}
	}

	c.JSON(http.StatusOK, gin.H{
		"runners": runners,
		"orgs":    orgs,
	})
}

func (s *service) getAllRunnerViews(ctx context.Context, orgID string) ([]views.AllRunnerView, error) {
	query := s.db.WithContext(ctx).
		Preload("RunnerGroup").
		Order("created_at DESC")

	if orgID != "" {
		query = query.Where(app.Runner{OrgID: orgID})
	}

	var runners []app.Runner
	if res := query.Find(&runners); res.Error != nil {
		return nil, res.Error
	}

	// batch-collect org IDs and install owner IDs for lookups
	orgIDs := make(map[string]bool)
	installIDs := make(map[string]bool)
	for _, r := range runners {
		orgIDs[r.OrgID] = true
		if r.RunnerGroup.OwnerType == "installs" {
			installIDs[r.RunnerGroup.OwnerID] = true
		}
	}

	// batch fetch org names
	orgNames := make(map[string]string)
	if len(orgIDs) > 0 {
		ids := mapKeys(orgIDs)
		var orgs []app.Org
		s.db.WithContext(ctx).Select("id", "name").Where("id IN ?", ids).Find(&orgs)
		for _, o := range orgs {
			orgNames[o.ID] = o.Name
		}
	}

	// batch fetch install names
	installNames := make(map[string]string)
	if len(installIDs) > 0 {
		ids := mapKeys(installIDs)
		var installs []app.Install
		s.db.WithContext(ctx).Select("id", "name").Where("id IN ?", ids).Find(&installs)
		for _, i := range installs {
			installNames[i.ID] = i.Name
		}
	}

	// batch fetch latest process per runner
	runnerIDs := make([]string, len(runners))
	for i, r := range runners {
		runnerIDs[i] = r.ID
	}

	type processInfo struct {
		Online      bool
		Version     string
		ProcessType string
	}
	processMap := make(map[string]processInfo)
	if len(runnerIDs) > 0 {
		var processes []app.RunnerProcess
		// use a subquery to get the latest process per runner
		s.db.WithContext(ctx).
			Raw(`SELECT DISTINCT ON (runner_id) * FROM runner_processes
				 WHERE runner_id IN ? AND deleted_at = 0
				 ORDER BY runner_id, created_at DESC`, runnerIDs).
			Scan(&processes)
		for _, p := range processes {
			processMap[p.RunnerID] = processInfo{
				Online:      p.ProcessStatus() == app.RunnerProcessStatusActive,
				Version:     p.Version,
				ProcessType: string(p.Type),
			}
		}
	}

	result := make([]views.AllRunnerView, 0, len(runners))
	for _, runner := range runners {
		view := views.AllRunnerView{
			Runner:    runner,
			OrgName:   orgNames[runner.OrgID],
			GroupType: string(runner.RunnerGroup.Type),
		}

		if runner.RunnerGroup.OwnerType == "installs" {
			view.InstallID = runner.RunnerGroup.OwnerID
			view.InstallName = installNames[runner.RunnerGroup.OwnerID]
		}

		if pi, ok := processMap[runner.ID]; ok {
			view.ProcessOnline = pi.Online
			view.Version = pi.Version
			view.ProcessType = pi.ProcessType
		}

		result = append(result, view)
	}

	return result, nil
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
