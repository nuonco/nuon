package service

import (
	"context"
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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
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
// @Param					component_id	query	string	false	"compare rollout on one component's applied config instead of the whole install"
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

func (s *service) getAppFleet(ctx context.Context, orgID, appID string, f fleetFilters, perGroup int) (*AppFleetResponse, error) {
	var resp AppFleetResponse
	err := s.fleetTx(ctx, func(tx *gorm.DB) error {
		branches, err := loadFleetBranches(tx, orgID, appID)
		if err != nil {
			return err
		}
		targetsJSON, err := fleetTargetsJSON(branches.targets)
		if err != nil {
			return err
		}
		configs, err := fleetGroupConfigs(tx, orgID, appID)
		if err != nil {
			return err
		}

		view := fleetViewName(tx)

		countsSQL, countsArgs, err := fleetCountsSQL(view, targetsJSON, orgID, appID, f)
		if err != nil {
			return err
		}
		var counts []fleetCountRow
		if err := tx.Raw(countsSQL, countsArgs...).Scan(&counts).Error; err != nil {
			return fmt.Errorf("unable to count installs: %w", err)
		}

		topSQL, topArgs, err := fleetTopPerGroupSQL(view, targetsJSON, orgID, appID, f, perGroup)
		if err != nil {
			return err
		}
		var rows []fleetInstallRow
		if err := tx.Raw(topSQL, topArgs...).Scan(&rows).Error; err != nil {
			return fmt.Errorf("unable to list fleet installs: %w", err)
		}

		resp = buildAppFleet(counts, rows, branches, configs)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type fleetBranchRow struct {
	AppBranchID string `gorm:"column:app_branch_id"`
	Name        string `gorm:"column:name"`
	RunID       string `gorm:"column:run_id"`
	AppConfigID string `gorm:"column:app_config_id"`
}

type fleetBranches struct {
	names   map[string]string
	targets map[string]AppFleetTargetRun
}

func loadFleetBranches(tx *gorm.DB, orgID, appID string) (fleetBranches, error) {
	var rows []fleetBranchRow
	err := tx.Raw(`SELECT b.id AS app_branch_id, b.name, COALESCE(r.id, '') AS run_id, COALESCE(r.app_config_id, '') AS app_config_id
FROM app_branches b
LEFT JOIN LATERAL (
  SELECT r.id, r.app_config_id
  FROM app_branch_runs r
  WHERE r.app_branch_id = b.id
    AND r.deleted_at = 0
    AND r.app_config_id <> ''
    AND r.no_config_changes = false
    AND r.plan_only = false
    AND r.run_type IS DISTINCT FROM ?
    AND NOT EXISTS (
      SELECT 1 FROM app_branch_run_previews p
      WHERE p.app_branch_run_id = r.id AND p.deleted_at = 0
    )
  ORDER BY r.created_at DESC, r.id DESC
  LIMIT 1
) r ON true
WHERE b.org_id = ? AND b.app_id = ? AND b.deleted_at = 0`,
		string(app.AppBranchRunTypeGitPreview), orgID, appID).Scan(&rows).Error
	if err != nil {
		return fleetBranches{}, fmt.Errorf("unable to get branch target runs: %w", err)
	}

	out := fleetBranches{
		names:   make(map[string]string, len(rows)),
		targets: make(map[string]AppFleetTargetRun, len(rows)),
	}
	for _, row := range rows {
		out.names[row.AppBranchID] = row.Name
		if row.RunID != "" {
			out.targets[row.AppBranchID] = AppFleetTargetRun{ID: row.RunID, AppConfigID: row.AppConfigID}
		}
	}
	return out, nil
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

func fleetGroupConfigs(tx *gorm.DB, orgID, appID string) (map[string]map[string]fleetGroupConfig, error) {
	var latestIDs []string
	err := tx.Raw(`SELECT c.id
FROM app_branches b
CROSS JOIN LATERAL (
  SELECT c.id
  FROM app_branch_configs c
  WHERE c.app_branch_id = b.id AND c.deleted_at = 0
  ORDER BY c.created_at DESC, c.id DESC
  LIMIT 1
) c
WHERE b.org_id = ? AND b.app_id = ? AND b.deleted_at = 0`, orgID, appID).Scan(&latestIDs).Error
	if err != nil {
		return nil, fmt.Errorf("unable to get latest branch configs: %w", err)
	}
	if len(latestIDs) == 0 {
		return map[string]map[string]fleetGroupConfig{}, nil
	}

	var configs []app.AppBranchConfig
	err = tx.
		Scopes(scopes.WithDisableViews).
		Select("id", "app_branch_id").
		Preload("InstallGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("id IN ?", latestIDs).
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
	counts []fleetCountRow,
	rows []fleetInstallRow,
	branches fleetBranches,
	configs map[string]map[string]fleetGroupConfig,
) AppFleetResponse {
	b := &fleetBuilder{
		branches: map[string]*AppFleetBranch{},
		groups:   map[fleetGroupKey]*AppFleetGroup{},
	}

	for _, c := range counts {
		branch, g := b.ensure(c.AppBranchID, c.AppBranchGroup, branches.names[c.AppBranchID])
		g.Counts.Total += c.Total
		branch.Counts.Total += c.Total
		b.summary.Total += c.Total
		g.Counts.addMatched(c.Rollout, c.Matched)
		branch.Counts.addMatched(c.Rollout, c.Matched)
		b.summary.addMatched(c.Rollout, c.Matched)
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

	out := make([]*AppFleetBranch, 0, len(b.branches))
	for key, branch := range b.branches {
		if target, ok := branches.targets[key]; ok {
			branch.TargetRun = &target
		}
		sort.SliceStable(branch.Groups, func(i, j int) bool {
			return fleetGroupLess(branch.Groups[i], branch.Groups[j])
		})
		out = append(out, branch)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return fleetBranchLess(out[i], out[j])
	})

	return AppFleetResponse{Summary: b.summary, Branches: out}
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
