package service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type AppFleetTargetRun struct {
	ID          string `json:"id"`
	AppConfigID string `json:"app_config_id"`
}

type AppFleetCounts struct {
	Total   int64 `json:"total"`
	Matched int64 `json:"matched"`
	Applied int64 `json:"applied"`
	Behind  int64 `json:"behind"`
	Unknown int64 `json:"unknown"`
}

type AppFleetSummary struct {
	AppFleetCounts
	RunnerNotActive int64 `json:"runner_not_active"`
}

type AppFleetInstall struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	AppBranchID     *string       `json:"app_branch_id"`
	AppBranchGroup  *string       `json:"app_branch_group"`
	Rollout         string        `json:"rollout"`
	Health          string        `json:"health"`
	RunnerStatus    string        `json:"runner_status"`
	AppConfigID     string        `json:"app_config_id"`
	AppliedConfigID string        `json:"applied_config_id"`
	Labels          labels.Labels `json:"labels"`
}

type AppFleetGroup struct {
	Name        *string           `json:"name"`
	Order       *int              `json:"order"`
	MatchLabels labels.Labels     `json:"match_labels"`
	Counts      AppFleetCounts    `json:"counts"`
	Installs    []AppFleetInstall `json:"installs"`
	HasMore     bool              `json:"has_more"`
}

type AppFleetBranch struct {
	ID        *string            `json:"id"`
	Name      *string            `json:"name"`
	TargetRun *AppFleetTargetRun `json:"target_run"`
	Counts    AppFleetCounts     `json:"counts"`
	Groups    []*AppFleetGroup   `json:"groups"`
}

type AppFleetResponse struct {
	Summary  AppFleetSummary   `json:"summary"`
	Branches []*AppFleetBranch `json:"branches"`
}

// @ID						GetAppFleet
// @Summary				app fleet summary, branches, and install groups
// @Description			Returns the app's installs grouped by app branch and install group, with counts per lane and the first installs of each group. Filters apply to counts and to the returned installs. Rollout compares each install's applied app config against the target run of its branch.
// @Param					app_id			path	string	true	"app ID"
// @Param					q				query	string	false	"search by install name or ID"
// @Param					labels			query	string	false	"label filter (key:value,key:value; value * matches key presence)"
// @Param					health			query	string	false	"comma-separated health: unhealthy,degraded,unknown,progressing,healthy,no-signal"
// @Param					runner_status	query	string	false	"runner filter: active or not-active"
// @Param					rollout			query	string	false	"rollout filter: applied, behind, or unknown"
// @Param					app_branch_id	query	string	false	"limit to one branch; __none__ for installs without a branch"
// @Param					app_branch_group	query	string	false	"limit to one install group; __none__ for installs without a group"
// @Param					per_group		query	int		false	"installs returned per group"	Default(10)
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	service.AppFleetResponse
// @Router					/v1/apps/{app_id}/fleet [GET]
func (s *service) GetAppFleet(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	currentApp, err := s.findAppByNameOrID(ctx, org.ID, ctx.Param("app_id"))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	filters, err := parseFleetFilters(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	perGroup, err := parseFleetPerGroup(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	resp, err := s.getAppFleet(ctx, org.ID, currentApp.ID, filters, perGroup)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app fleet: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (s *service) getAppFleet(ctx *gin.Context, orgID, appID string, f fleetFilters, perGroup int) (*AppFleetResponse, error) {
	targets, err := s.fleetTargets(ctx, orgID, appID)
	if err != nil {
		return nil, err
	}
	targetsJSON, err := fleetTargetsJSON(targets)
	if err != nil {
		return nil, err
	}
	configs, err := s.fleetGroupConfigs(ctx, orgID, appID)
	if err != nil {
		return nil, err
	}

	view := fleetViewName(s.db)

	var totals []fleetTotalRow
	if err := s.db.WithContext(ctx).Raw(fleetTotalsSQL(view), orgID, appID).Scan(&totals).Error; err != nil {
		return nil, fmt.Errorf("unable to count installs: %w", err)
	}

	countsSQL, countsArgs, err := fleetCountsSQL(view, targetsJSON, orgID, appID, f)
	if err != nil {
		return nil, err
	}
	var counts []fleetCountRow
	if err := s.db.WithContext(ctx).Raw(countsSQL, countsArgs...).Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("unable to count matching installs: %w", err)
	}

	topSQL, topArgs, err := fleetTopPerGroupSQL(view, targetsJSON, orgID, appID, f, perGroup)
	if err != nil {
		return nil, err
	}
	var rows []fleetInstallRow
	if err := s.db.WithContext(ctx).Raw(topSQL, topArgs...).Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to list fleet installs: %w", err)
	}

	resp := buildAppFleet(totals, counts, rows, targets, configs)
	return &resp, nil
}

type fleetTargetRow struct {
	AppBranchID string `gorm:"column:app_branch_id"`
	RunID       string `gorm:"column:run_id"`
	AppConfigID string `gorm:"column:app_config_id"`
}

func (s *service) fleetTargets(ctx *gin.Context, orgID, appID string) (map[string]AppFleetTargetRun, error) {
	var rows []fleetTargetRow
	err := s.db.WithContext(ctx).Raw(`SELECT DISTINCT ON (r.app_branch_id) r.app_branch_id, r.id AS run_id, r.app_config_id
FROM app_branch_runs r
WHERE r.org_id = ?
  AND r.app_config_id <> ''
  AND r.no_config_changes = false
  AND r.plan_only = false
  AND r.run_type IS DISTINCT FROM ?
  AND r.deleted_at = 0
  AND NOT EXISTS (
    SELECT 1 FROM app_branch_run_previews p
    WHERE p.app_branch_run_id = r.id AND p.deleted_at = 0
  )
  AND r.app_branch_id IN (
    SELECT id FROM app_branches WHERE org_id = ? AND app_id = ? AND deleted_at = 0
  )
ORDER BY r.app_branch_id, r.created_at DESC, r.id DESC`,
		orgID, string(app.AppBranchRunTypeGitPreview), orgID, appID).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get branch target runs: %w", err)
	}

	targets := make(map[string]AppFleetTargetRun, len(rows))
	for _, row := range rows {
		targets[row.AppBranchID] = AppFleetTargetRun{ID: row.RunID, AppConfigID: row.AppConfigID}
	}
	return targets, nil
}

func fleetTargetsJSON(targets map[string]AppFleetTargetRun) (string, error) {
	configs := make(map[string]string, len(targets))
	for branchID, target := range targets {
		configs[branchID] = target.AppConfigID
	}
	raw, err := json.Marshal(configs)
	if err != nil {
		return "", fmt.Errorf("unable to encode target configs: %w", err)
	}
	return string(raw), nil
}

type fleetGroupConfig struct {
	order       int
	matchLabels labels.Labels
}

func (s *service) fleetGroupConfigs(ctx *gin.Context, orgID, appID string) (map[string]map[string]fleetGroupConfig, error) {
	branches := s.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Select("id").
		Where("org_id = ? AND app_id = ?", orgID, appID)

	latest := s.db.WithContext(ctx).
		Model(&app.AppBranchConfig{}).
		Select("DISTINCT ON (app_branch_id) id").
		Where("org_id = ? AND app_branch_id IN (?)", orgID, branches).
		Order("app_branch_id, created_at DESC")

	var configs []app.AppBranchConfig
	err := s.db.WithContext(ctx).
		Preload("InstallGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("id IN (?)", latest).
		Find(&configs).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get branch install groups: %w", err)
	}

	out := make(map[string]map[string]fleetGroupConfig, len(configs))
	for _, cfg := range configs {
		groups := make(map[string]fleetGroupConfig, len(cfg.InstallGroups))
		for _, group := range cfg.InstallGroups {
			var match labels.Labels
			if group.LabelSelector != nil {
				match = group.LabelSelector.MatchLabels
			}
			groups[group.Name] = fleetGroupConfig{order: group.Order, matchLabels: match}
		}
		out[cfg.AppBranchID] = groups
	}
	return out, nil
}

type fleetGroupKey struct {
	branchID string
	group    string
}

type fleetBuilder struct {
	summary  AppFleetSummary
	branches map[string]*AppFleetBranch
	groups   map[fleetGroupKey]*AppFleetGroup
}

func (b *fleetBuilder) ensure(branchID, group, branchName string) (*AppFleetBranch, *AppFleetGroup) {
	branch, ok := b.branches[branchID]
	if !ok {
		branch = &AppFleetBranch{ID: nullableString(branchID), Groups: []*AppFleetGroup{}}
		b.branches[branchID] = branch
	}
	if branch.Name == nil {
		branch.Name = nullableString(branchName)
	}

	key := fleetGroupKey{branchID: branchID, group: group}
	g, ok := b.groups[key]
	if !ok {
		g = &AppFleetGroup{Name: nullableString(group), Installs: []AppFleetInstall{}}
		b.groups[key] = g
		branch.Groups = append(branch.Groups, g)
	}
	return branch, g
}

func (c *AppFleetCounts) addMatched(rollout string, n int64) {
	c.Matched += n
	switch rollout {
	case "applied":
		c.Applied += n
	case "behind":
		c.Behind += n
	default:
		c.Unknown += n
	}
}

func buildAppFleet(
	totals []fleetTotalRow,
	counts []fleetCountRow,
	rows []fleetInstallRow,
	targets map[string]AppFleetTargetRun,
	configs map[string]map[string]fleetGroupConfig,
) AppFleetResponse {
	b := &fleetBuilder{
		branches: map[string]*AppFleetBranch{},
		groups:   map[fleetGroupKey]*AppFleetGroup{},
	}

	for _, t := range totals {
		branch, g := b.ensure(t.AppBranchID, t.AppBranchGroup, t.AppBranchName)
		g.Counts.Total += t.Total
		branch.Counts.Total += t.Total
		b.summary.Total += t.Total
	}

	for _, c := range counts {
		branch, g := b.ensure(c.AppBranchID, c.AppBranchGroup, "")
		g.Counts.addMatched(c.Rollout, c.Installs)
		branch.Counts.addMatched(c.Rollout, c.Installs)
		b.summary.addMatched(c.Rollout, c.Installs)
		b.summary.RunnerNotActive += c.RunnerNotActive
	}

	for _, r := range rows {
		_, g := b.ensure(r.AppBranchID, r.AppBranchGroup, "")
		g.Installs = append(g.Installs, fleetInstallFromRow(r))
	}

	for key, g := range b.groups {
		g.HasMore = g.Counts.Matched > int64(len(g.Installs))
		if cfg, ok := configs[key.branchID][key.group]; ok {
			order := cfg.order
			g.Order = &order
			g.MatchLabels = cfg.matchLabels
		}
	}

	branches := make([]*AppFleetBranch, 0, len(b.branches))
	for key, branch := range b.branches {
		if target, ok := targets[key]; ok {
			branch.TargetRun = &target
		}
		sort.SliceStable(branch.Groups, func(i, j int) bool {
			return fleetGroupLess(branch.Groups[i], branch.Groups[j])
		})
		branches = append(branches, branch)
	}
	sort.SliceStable(branches, func(i, j int) bool {
		return fleetBranchLess(branches[i], branches[j])
	})

	return AppFleetResponse{Summary: b.summary, Branches: branches}
}

func (s *AppFleetSummary) addMatched(rollout string, n int64) {
	s.AppFleetCounts.addMatched(rollout, n)
}

func fleetInstallFromRow(r fleetInstallRow) AppFleetInstall {
	health := r.CompositeHealthStatus
	if health == "" {
		health = fleetNoSignalToken
	}
	installLabels := r.Labels
	if installLabels == nil {
		installLabels = labels.Labels{}
	}
	return AppFleetInstall{
		ID:              r.ID,
		Name:            r.Name,
		AppBranchID:     nullableString(r.AppBranchID),
		AppBranchGroup:  nullableString(r.AppBranchGroup),
		Rollout:         r.Rollout,
		Health:          health,
		RunnerStatus:    r.RunnerStatus,
		AppConfigID:     r.AppConfigID,
		AppliedConfigID: r.AppliedConfigID,
		Labels:          installLabels,
	}
}

func fleetGroupOrder(g *AppFleetGroup) int {
	if g.Order == nil {
		return math.MaxInt
	}
	return *g.Order
}

func fleetGroupLess(a, b *AppFleetGroup) bool {
	if (a.Name == nil) != (b.Name == nil) {
		return b.Name == nil
	}
	if ao, bo := fleetGroupOrder(a), fleetGroupOrder(b); ao != bo {
		return ao < bo
	}
	return derefString(a.Name) < derefString(b.Name)
}

func fleetBranchLess(a, b *AppFleetBranch) bool {
	if (a.ID == nil) != (b.ID == nil) {
		return b.ID == nil
	}
	return derefString(a.Name) < derefString(b.Name)
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
