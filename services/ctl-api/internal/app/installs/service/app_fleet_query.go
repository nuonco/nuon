package service

import (
	"context"
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

	fleetStatementTimeout = "5s"
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
	componentID  string
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
		componentID:  strings.TrimSpace(ctx.Query("component_id")),
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

func (s *service) fleetTx(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(fmt.Sprintf("SET LOCAL statement_timeout = '%s'", fleetStatementTimeout)).Error; err != nil {
			return fmt.Errorf("unable to set fleet statement timeout: %w", err)
		}
		// The per-install laterals push the cost estimate past jit_above_cost, and
		// compiling costs more than these short index-bound queries take to run.
		if err := tx.Exec("SET LOCAL jit = off").Error; err != nil {
			return fmt.Errorf("unable to disable jit for fleet queries: %w", err)
		}
		return fn(tx)
	})
}

func fleetRolloutSQL(view string, f fleetFilters) string {
	applied := "f.applied_config_id"
	component := ""
	if f.componentID != "" {
		applied = "COALESCE(ic.applied_config_id, '')"
		component = `LEFT JOIN LATERAL (
	SELECT c.app_config_ref->>'applied_config_id' AS applied_config_id
	FROM install_components c
	WHERE c.install_id = f.id AND c.component_id = ? AND c.deleted_at = 0
	LIMIT 1
) ic ON true
`
	}
	return fmt.Sprintf(`SELECT f.*,
	CASE
		WHEN %[2]s = '' OR t.target_config_id IS NULL THEN 'unknown'
		WHEN %[2]s = t.target_config_id THEN 'applied'
		ELSE 'behind'
	END AS rollout
FROM %[1]s f
CROSS JOIN LATERAL (SELECT (?::jsonb) ->> f.app_branch_id AS target_config_id) t
%[3]sWHERE f.org_id = ? AND f.app_id = ?`, view, applied, component)
}

func fleetRolloutArgs(targetsJSON, orgID, appID string, f fleetFilters) []any {
	if f.componentID != "" {
		return []any{targetsJSON, f.componentID, orgID, appID}
	}
	return []any{targetsJSON, orgID, appID}
}

func fleetScopedSQL(view, targetsJSON, orgID, appID string, f fleetFilters) (string, []any, error) {
	where, args, err := f.where()
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("SELECT fleet.* FROM (%s) fleet WHERE %s", fleetRolloutSQL(view, f), where),
		append(fleetRolloutArgs(targetsJSON, orgID, appID, f), args...), nil
}

type fleetCountRow struct {
	AppBranchID     string `gorm:"column:app_branch_id"`
	AppBranchGroup  string `gorm:"column:app_branch_group"`
	Rollout         string `gorm:"column:rollout"`
	Total           int64  `gorm:"column:total"`
	Matched         int64  `gorm:"column:matched"`
	RunnerNotActive int64  `gorm:"column:runner_not_active"`
}

func fleetCountsSQL(view, targetsJSON, orgID, appID string, f fleetFilters) (string, []any, error) {
	where, whereArgs, err := f.where()
	if err != nil {
		return "", nil, err
	}
	sql := fmt.Sprintf(`SELECT fleet.app_branch_id, fleet.app_branch_group, fleet.rollout,
	COUNT(*) AS total,
	COUNT(*) FILTER (WHERE %s) AS matched,
	COUNT(*) FILTER (WHERE (%s) AND fleet.runner_status <> 'active') AS runner_not_active
FROM (%s) fleet
GROUP BY fleet.app_branch_id, fleet.app_branch_group, fleet.rollout`, where, where, fleetRolloutSQL(view, f))

	args := make([]any, 0, len(whereArgs)*2+4)
	args = append(args, whereArgs...)
	args = append(args, whereArgs...)
	args = append(args, fleetRolloutArgs(targetsJSON, orgID, appID, f)...)
	return sql, args, nil
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
