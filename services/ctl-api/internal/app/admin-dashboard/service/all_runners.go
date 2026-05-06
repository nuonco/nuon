package service

import (
	"context"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

const allRunnersPerPage = 50

type runnerCategoryCount struct {
	Label string `json:"label" gorm:"column:label"`
	Value int    `json:"value" gorm:"column:value"`
}

type runnerStats struct {
	GroupType   []runnerCategoryCount `json:"group_type"`
	Version     []runnerCategoryCount `json:"version"`
	ProcessType []runnerCategoryCount `json:"process_type"`
}

func (s *service) AllRunners(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Query("org_id")
	page := getPageFromQuery(c)

	runners, totalCount, err := s.getAllRunnerViews(ctx, orgID, page)
	if err != nil {
		s.l.Error("failed to get all runners", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch runners"})
		return
	}

	stats, err := s.getAllRunnerStats(ctx, orgID)
	if err != nil {
		s.l.Warn("failed to get runner stats", zap.Error(err))
		stats = runnerStats{}
	}

	orgs := s.getOrgOptions(ctx)
	if orgs == nil {
		orgs = []views.OrgOption{}
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(allRunnersPerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"runners":     runners,
		"orgs":        orgs,
		"stats":       stats,
		"page":        page,
		"total_pages": totalPages,
		"total_count": totalCount,
	})
}

func (s *service) getAllRunnerViews(ctx context.Context, orgID string, page int) ([]views.AllRunnerView, int64, error) {
	query := s.db.WithContext(ctx).Model(&app.Runner{})
	if orgID != "" {
		query = query.Where(app.Runner{OrgID: orgID})
	}

	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * allRunnersPerPage

	var runners []app.Runner
	if res := query.
		Preload("RunnerGroup").
		Order("created_at DESC").
		Limit(allRunnersPerPage).
		Offset(offset).
		Find(&runners); res.Error != nil {
		return nil, 0, res.Error
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

	return result, totalCount, nil
}

// getAllRunnerStats returns cluster-wide runner aggregations for the pie charts.
// Each query returns one row per category, so memory cost is bounded by the
// number of distinct group types / versions / process types.
func (s *service) getAllRunnerStats(ctx context.Context, orgID string) (runnerStats, error) {
	var stats runnerStats

	// group_type: counts by runner_groups.type
	groupTypeQuery := s.db.WithContext(ctx).
		Table("runners").
		Select("COALESCE(NULLIF(runner_groups.type, ''), 'unknown') AS label, COUNT(*) AS value").
		Joins("JOIN runner_groups ON runner_groups.id = runners.runner_group_id").
		Where("runners.deleted_at = 0").
		Group("runner_groups.type")
	if orgID != "" {
		groupTypeQuery = groupTypeQuery.Where("runners.org_id = ?", orgID)
	}
	if err := groupTypeQuery.Scan(&stats.GroupType).Error; err != nil {
		return stats, err
	}

	// version + process_type: counts over the latest runner_process per runner.
	// Runners with no process row count as 'unknown'/'none' to match the prior
	// per-runner behavior.
	latestProcessSQL := `
		WITH latest AS (
			SELECT DISTINCT ON (runner_id) runner_id, version, type
			FROM runner_processes
			WHERE deleted_at = 0
			ORDER BY runner_id, created_at DESC
		)
	`
	args := []any{}
	scope := "WHERE r.deleted_at = 0"
	if orgID != "" {
		scope += " AND r.org_id = ?"
		args = append(args, orgID)
	}

	versionSQL := latestProcessSQL + `
		SELECT COALESCE(NULLIF(latest.version, ''), 'unknown') AS label, COUNT(*) AS value
		FROM runners r
		LEFT JOIN latest ON latest.runner_id = r.id
		` + scope + `
		GROUP BY 1
	`
	if err := s.db.WithContext(ctx).Raw(versionSQL, args...).Scan(&stats.Version).Error; err != nil {
		return stats, err
	}

	processTypeSQL := latestProcessSQL + `
		SELECT COALESCE(NULLIF(latest.type, ''), 'none') AS label, COUNT(*) AS value
		FROM runners r
		LEFT JOIN latest ON latest.runner_id = r.id
		` + scope + `
		GROUP BY 1
	`
	if err := s.db.WithContext(ctx).Raw(processTypeSQL, args...).Scan(&stats.ProcessType).Error; err != nil {
		return stats, err
	}

	return stats, nil
}

func mapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
