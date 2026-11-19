package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
)

const (
	fleetNoneToken       = "__none__"
	fleetNoSignalToken   = "no-signal"
	fleetRunnerActive    = "active"
	fleetRunnerNotActive = "not-active"
	fleetDefaultPerGroup = 10
	fleetMaxPerGroup     = 50
)

var (
	fleetHealthValues = map[string]bool{
		"unhealthy":        true,
		"degraded":         true,
		"unknown":          true,
		"progressing":      true,
		"healthy":          true,
		fleetNoSignalToken: true,
	}
	fleetRolloutValues = map[string]bool{
		"applied": true,
		"behind":  true,
		"unknown": true,
	}
)

const fleetOrderBy = `CASE fleet.rollout WHEN 'behind' THEN 0 WHEN 'unknown' THEN 1 ELSE 2 END,
	CASE fleet.composite_health_status
		WHEN 'unhealthy' THEN 0
		WHEN 'degraded' THEN 1
		WHEN 'unknown' THEN 2
		WHEN 'progressing' THEN 3
		WHEN 'healthy' THEN 4
		ELSE 5
	END,
	fleet.name, fleet.id`

type fleetFilters struct {
	search       string
	labels       labels.Labels
	health       []string
	runnerStatus string
	rollout      string
	branchID     *string
	group        *string
}

func parseFleetFilters(ctx *gin.Context) (fleetFilters, error) {
	f := fleetFilters{
		search:       strings.TrimSpace(ctx.Query("q")),
		labels:       labels.ParseLabelsQuery(ctx.Query("labels")),
		health:       parseCommaSeparated(ctx.Query("health")),
		runnerStatus: ctx.Query("runner_status"),
		rollout:      ctx.Query("rollout"),
	}

	for _, h := range f.health {
		if !fleetHealthValues[h] {
			return f, stderr.ErrUser{
				Err:         fmt.Errorf("invalid health %q", h),
				Description: "health must be one of: unhealthy, degraded, unknown, progressing, healthy, no-signal",
			}
		}
	}

	switch f.runnerStatus {
	case "", fleetRunnerActive, fleetRunnerNotActive:
	default:
		return f, stderr.ErrUser{
			Err:         fmt.Errorf("invalid runner_status %q", f.runnerStatus),
			Description: "runner_status must be one of: active, not-active",
		}
	}

	if f.rollout != "" && !fleetRolloutValues[f.rollout] {
		return f, stderr.ErrUser{
			Err:         fmt.Errorf("invalid rollout %q", f.rollout),
			Description: "rollout must be one of: applied, behind, unknown",
		}
	}

	if v := ctx.Query("app_branch_id"); v != "" {
		f.branchID = &v
	}
	if v := ctx.Query("app_branch_group"); v != "" {
		f.group = &v
	}

	return f, nil
}

func parseFleetPerGroup(ctx *gin.Context) (int, error) {
	raw := ctx.Query("per_group")
	if raw == "" {
		return fleetDefaultPerGroup, nil
	}

	n, err := strconv.Atoi(raw)
	if err != nil || n < 1 || n > fleetMaxPerGroup {
		return 0, stderr.ErrUser{
			Err:         fmt.Errorf("invalid per_group %q", raw),
			Description: fmt.Sprintf("per_group must be between 1 and %d", fleetMaxPerGroup),
		}
	}
	return n, nil
}

func fleetColumnValue(v string) string {
	if v == fleetNoneToken {
		return ""
	}
	return v
}

func (f fleetFilters) where() (string, []any, error) {
	var clauses []string
	var args []any

	if f.search != "" {
		pattern := "%" + f.search + "%"
		clauses = append(clauses, "(fleet.name ILIKE ? OR fleet.id ILIKE ?)")
		args = append(args, pattern, pattern)
	}

	exact := labels.Labels{}
	for key, value := range f.labels {
		if value == "*" {
			clauses = append(clauses, "jsonb_exists(fleet.labels, ?)")
			args = append(args, key)
			continue
		}
		exact[key] = value
	}
	if len(exact) > 0 {
		raw, err := json.Marshal(exact)
		if err != nil {
			return "", nil, fmt.Errorf("unable to encode labels filter: %w", err)
		}
		clauses = append(clauses, "fleet.labels @> ?::jsonb")
		args = append(args, string(raw))
	}

	if len(f.health) > 0 {
		values := make([]string, 0, len(f.health))
		for _, h := range f.health {
			if h == fleetNoSignalToken {
				h = ""
			}
			values = append(values, h)
		}
		clauses = append(clauses, "fleet.composite_health_status IN ?")
		args = append(args, values)
	}

	switch f.runnerStatus {
	case fleetRunnerActive:
		clauses = append(clauses, "fleet.runner_status = ?")
		args = append(args, fleetRunnerActive)
	case fleetRunnerNotActive:
		clauses = append(clauses, "fleet.runner_status <> ?")
		args = append(args, fleetRunnerActive)
	}

	if f.rollout != "" {
		clauses = append(clauses, "fleet.rollout = ?")
		args = append(args, f.rollout)
	}

	if f.branchID != nil {
		clauses = append(clauses, "fleet.app_branch_id = ?")
		args = append(args, fleetColumnValue(*f.branchID))
	}

	if f.group != nil {
		clauses = append(clauses, "fleet.app_branch_group = ?")
		args = append(args, fleetColumnValue(*f.group))
	}

	if len(clauses) == 0 {
		return "TRUE", nil, nil
	}
	return strings.Join(clauses, " AND "), args, nil
}

func fleetViewName(db *gorm.DB) string {
	return views.CustomViewName(db, &app.Install{}, "fleet_view_v1")
}

// fleetRolloutSQL derives each install's rollout from its applied config and
// the target config for its branch. The target map is passed as one jsonb
// argument so the branch lookup is a single keyed lookup per row.
func fleetRolloutSQL(view string) string {
	return fmt.Sprintf(`SELECT f.*,
	CASE
		WHEN f.applied_config_id = '' OR t.target_config_id IS NULL THEN 'unknown'
		WHEN f.applied_config_id = t.target_config_id THEN 'applied'
		ELSE 'behind'
	END AS rollout
FROM %s f
LEFT JOIN LATERAL (
	SELECT value AS target_config_id
	FROM jsonb_each_text(?::jsonb)
	WHERE key = f.app_branch_id
) t ON true
WHERE f.org_id = ? AND f.app_id = ?`, view)
}

func fleetScopedSQL(view, targetsJSON, orgID, appID string, f fleetFilters) (string, []any, error) {
	where, args, err := f.where()
	if err != nil {
		return "", nil, err
	}
	base := []any{targetsJSON, orgID, appID}
	return fmt.Sprintf("SELECT fleet.* FROM (%s) fleet WHERE %s", fleetRolloutSQL(view), where),
		append(base, args...), nil
}

type fleetTotalRow struct {
	AppBranchID    string `gorm:"column:app_branch_id"`
	AppBranchGroup string `gorm:"column:app_branch_group"`
	AppBranchName  string `gorm:"column:app_branch_name"`
	Total          int64  `gorm:"column:total"`
}

func fleetTotalsSQL(view string) string {
	return fmt.Sprintf(`SELECT app_branch_id, app_branch_group, MAX(app_branch_name) AS app_branch_name, COUNT(*) AS total
FROM %s
WHERE org_id = ? AND app_id = ?
GROUP BY app_branch_id, app_branch_group`, view)
}

type fleetCountRow struct {
	AppBranchID     string `gorm:"column:app_branch_id"`
	AppBranchGroup  string `gorm:"column:app_branch_group"`
	Rollout         string `gorm:"column:rollout"`
	Installs        int64  `gorm:"column:installs"`
	RunnerNotActive int64  `gorm:"column:runner_not_active"`
}

func fleetCountsSQL(view, targetsJSON, orgID, appID string, f fleetFilters) (string, []any, error) {
	scoped, args, err := fleetScopedSQL(view, targetsJSON, orgID, appID, f)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf(`SELECT scoped.app_branch_id, scoped.app_branch_group, scoped.rollout,
	COUNT(*) AS installs,
	COUNT(*) FILTER (WHERE scoped.runner_status <> 'active') AS runner_not_active
FROM (%s) scoped
GROUP BY scoped.app_branch_id, scoped.app_branch_group, scoped.rollout`, scoped), args, nil
}

type fleetInstallRow struct {
	ID                    string        `gorm:"column:id"`
	Name                  string        `gorm:"column:name"`
	Labels                labels.Labels `gorm:"column:labels"`
	AppConfigID           string        `gorm:"column:app_config_id"`
	AppliedConfigID       string        `gorm:"column:applied_config_id"`
	AppBranchID           string        `gorm:"column:app_branch_id"`
	AppBranchGroup        string        `gorm:"column:app_branch_group"`
	CompositeHealthStatus string        `gorm:"column:composite_health_status"`
	RunnerStatus          string        `gorm:"column:runner_status"`
	Rollout               string        `gorm:"column:rollout"`
	GroupRank             int64         `gorm:"column:group_rank"`
}

func fleetTopPerGroupSQL(view, targetsJSON, orgID, appID string, f fleetFilters, perGroup int) (string, []any, error) {
	scoped, args, err := fleetScopedSQL(view, targetsJSON, orgID, appID, f)
	if err != nil {
		return "", nil, err
	}
	sql := fmt.Sprintf(`SELECT ranked.* FROM (
	SELECT fleet.*, ROW_NUMBER() OVER (
		PARTITION BY fleet.app_branch_id, fleet.app_branch_group
		ORDER BY %s
	) AS group_rank
	FROM (%s) fleet
) ranked
WHERE ranked.group_rank <= ?
ORDER BY ranked.app_branch_id, ranked.app_branch_group, ranked.group_rank`, fleetOrderBy, scoped)
	return sql, append(args, perGroup), nil
}

func fleetInstallsSQL(view, targetsJSON, orgID, appID string, f fleetFilters) (string, []any, error) {
	scoped, args, err := fleetScopedSQL(view, targetsJSON, orgID, appID, f)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("SELECT * FROM (%s) fleet ORDER BY %s LIMIT ? OFFSET ?", scoped, fleetOrderBy), args, nil
}

func nullableString(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
