package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// AppLabelKey represents a label key with its usage across entity types.
type AppLabelKey struct {
	Key         string   `json:"key"`
	Color       string   `json:"color"`
	IsOverride  bool     `json:"is_override"`
	Values      []string `json:"values"`
	EntityTypes []string `json:"entity_types"`
	UsageCount  int      `json:"usage_count"`
}

// AppLabelsResponse is the response for the app labels endpoint.
type AppLabelsResponse struct {
	Labels        []AppLabelKey     `json:"labels"`
	LabelColors   map[string]string `json:"label_colors"`
	DefaultColors []string          `json:"default_colors"`
}

// StandardLabelColors is a palette of 64 visually distinct colors for label defaults.
var StandardLabelColors = []string{
	"#2563eb", "#dc2626", "#16a34a", "#9333ea", "#ca8a04", "#0891b2",
	"#e11d48", "#4f46e5", "#059669", "#c026d3", "#d97706", "#0284c7",
	"#be123c", "#7c3aed", "#15803d", "#a21caf", "#b45309", "#0369a1",
	"#9f1239", "#6d28d9", "#166534", "#86198f", "#92400e", "#075985",
	"#881337", "#5b21b6", "#14532d", "#701a75", "#78350f", "#0c4a6e",
	"#6366f1", "#ef4444", "#22c55e", "#a855f7", "#eab308", "#06b6d4",
	"#f43f5e", "#818cf8", "#34d399", "#d946ef", "#f59e0b", "#0ea5e9",
	"#fb7185", "#a78bfa", "#4ade80", "#e879f9", "#fbbf24", "#38bdf8",
	"#f87171", "#c084fc", "#86efac", "#f0abfc", "#fcd34d", "#7dd3fc",
	"#fca5a5", "#d8b4fe", "#bbf7d0", "#f5d0fe", "#fde68a", "#bae6fd",
	"#3b82f6", "#f97316", "#8b5cf6", "#14b8a6",
}

// @ID						GetAppLabels
// @Summary				get all labels used across an app
// @Description			Returns all distinct label keys with values, usage counts, and assigned colors across components, actions, runbooks, and installs for an app.
// @Param					app_id	path	string	true	"app ID"
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
// @Success				200	{object}	AppLabelsResponse
// @Router					/v1/apps/{app_id}/labels [get]
func (s *service) GetAppLabels(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	app, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	labelKeys, err := s.getAppLabelKeys(ctx, org.ID, app.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app labels: %w", err))
		return
	}

	overrides := map[string]string{}
	if app.LabelColors != nil {
		overrides = map[string]string(app.LabelColors)
	}

	for i, lk := range labelKeys {
		if color, ok := overrides[lk.Key]; ok {
			labelKeys[i].Color = color
			labelKeys[i].IsOverride = true
		} else {
			labelKeys[i].Color = StandardLabelColors[i%len(StandardLabelColors)]
		}
	}

	ctx.JSON(http.StatusOK, AppLabelsResponse{
		Labels:        labelKeys,
		LabelColors:   overrides,
		DefaultColors: StandardLabelColors,
	})
}

type labelKeyRow struct {
	Key         string
	Value       string
	EntityType  string
	FirstUsedAt string
}

func (s *service) getAppLabelKeys(ctx context.Context, orgID, appID string) ([]AppLabelKey, error) {
	queries := []struct {
		entityType string
		sql        string
		args       []any
	}{
		{
			entityType: "component",
			sql:        "SELECT (jsonb_each_text(labels)).key AS key, (jsonb_each_text(labels)).value AS value, 'component' AS entity_type, created_at AS first_used_at FROM components WHERE labels IS NOT NULL AND labels != '{}'::jsonb AND deleted_at = 0 AND app_id = ?",
			args:       []any{appID},
		},
		{
			entityType: "action",
			sql:        "SELECT (jsonb_each_text(labels)).key AS key, (jsonb_each_text(labels)).value AS value, 'action' AS entity_type, created_at AS first_used_at FROM action_workflows WHERE labels IS NOT NULL AND labels != '{}'::jsonb AND deleted_at = 0 AND org_id = ? AND app_id = ?",
			args:       []any{orgID, appID},
		},
		{
			entityType: "runbook",
			sql:        "SELECT (jsonb_each_text(labels)).key AS key, (jsonb_each_text(labels)).value AS value, 'runbook' AS entity_type, created_at AS first_used_at FROM runbooks WHERE labels IS NOT NULL AND labels != '{}'::jsonb AND deleted_at = 0 AND org_id = ? AND app_id = ?",
			args:       []any{orgID, appID},
		},
		{
			entityType: "install",
			sql:        "SELECT (jsonb_each_text(labels)).key AS key, (jsonb_each_text(labels)).value AS value, 'install' AS entity_type, created_at AS first_used_at FROM installs WHERE labels IS NOT NULL AND labels != '{}'::jsonb AND deleted_at = 0 AND app_id = ?",
			args:       []any{appID},
		},
	}

	var allRows []labelKeyRow
	for _, q := range queries {
		var rows []labelKeyRow
		if err := s.db.WithContext(ctx).Raw(q.sql, q.args...).Scan(&rows).Error; err != nil {
			return nil, fmt.Errorf("unable to query labels from %s: %w", q.entityType, err)
		}
		allRows = append(allRows, rows...)
	}

	type keyInfo struct {
		values      map[string]bool
		entityTypes map[string]bool
		usageCount  int
		firstUsedAt string
	}

	keyMap := make(map[string]*keyInfo)
	var keyOrder []string

	for _, row := range allRows {
		info, exists := keyMap[row.Key]
		if !exists {
			info = &keyInfo{
				values:      make(map[string]bool),
				entityTypes: make(map[string]bool),
				firstUsedAt: row.FirstUsedAt,
			}
			keyMap[row.Key] = info
			keyOrder = append(keyOrder, row.Key)
		}
		info.values[row.Value] = true
		info.entityTypes[row.EntityType] = true
		info.usageCount++
		if row.FirstUsedAt < info.firstUsedAt {
			info.firstUsedAt = row.FirstUsedAt
		}
	}

	// Sort by first usage time
	sortByFirstUsed := make([]string, len(keyOrder))
	copy(sortByFirstUsed, keyOrder)
	for i := 0; i < len(sortByFirstUsed); i++ {
		for j := i + 1; j < len(sortByFirstUsed); j++ {
			if keyMap[sortByFirstUsed[i]].firstUsedAt > keyMap[sortByFirstUsed[j]].firstUsedAt {
				sortByFirstUsed[i], sortByFirstUsed[j] = sortByFirstUsed[j], sortByFirstUsed[i]
			}
		}
	}

	result := make([]AppLabelKey, 0, len(sortByFirstUsed))
	for _, key := range sortByFirstUsed {
		info := keyMap[key]
		values := make([]string, 0, len(info.values))
		for v := range info.values {
			values = append(values, v)
		}
		entityTypes := make([]string, 0, len(info.entityTypes))
		for et := range info.entityTypes {
			entityTypes = append(entityTypes, et)
		}
		result = append(result, AppLabelKey{
			Key:         key,
			Values:      values,
			EntityTypes: entityTypes,
			UsageCount:  info.usageCount,
		})
	}

	return result, nil
}
