package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type AppLabelKeySummary struct {
	Key          string   `json:"key"`
	Color        string   `json:"color"`
	DefaultColor string   `json:"default_color"`
	IsOverride   bool     `json:"is_override"`
	Values       []string `json:"values"`
	EntityTypes  []string `json:"entity_types"`
	UsageCount   int      `json:"usage_count"`
}

type AppLabelsResponse struct {
	Labels        []AppLabelKeySummary `json:"labels"`
	LabelColors   map[string]string    `json:"label_colors"`
	DefaultColors []string             `json:"default_colors"`
}

var StandardLabelColors = []string{
	"#2563eb", "#0d9488", "#16a34a", "#9333ea", "#ca8a04", "#0891b2",
	"#1d4ed8", "#4f46e5", "#059669", "#c026d3", "#d97706", "#0284c7",
	"#0f766e", "#7c3aed", "#15803d", "#a21caf", "#b45309", "#0369a1",
	"#115e59", "#6d28d9", "#166534", "#86198f", "#92400e", "#075985",
	"#134e4a", "#5b21b6", "#14532d", "#701a75", "#78350f", "#0c4a6e",
	"#6366f1", "#84cc16", "#22c55e", "#a855f7", "#eab308", "#06b6d4",
	"#65a30d", "#818cf8", "#34d399", "#d946ef", "#f59e0b", "#0ea5e9",
	"#2dd4bf", "#a78bfa", "#4ade80", "#e879f9", "#fbbf24", "#38bdf8",
	"#22d3ee", "#c084fc", "#86efac", "#f0abfc", "#fcd34d", "#7dd3fc",
	"#5eead4", "#d8b4fe", "#bbf7d0", "#f5d0fe", "#fde68a", "#bae6fd",
	"#3b82f6", "#f97316", "#8b5cf6", "#14b8a6",
}

// @ID						GetAppLabels
// @Summary				get all labels used across an app
// @Description			Returns all distinct label keys with values, usage counts, and assigned colors across components, actions, runbooks, and installs for an app.
// @Param					app_id	path	string	true	"app ID"
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
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

	currentApp, err := s.appByNameOrID(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	var rows []app.AppLabelKey
	res := s.db.WithContext(ctx).
		Where(app.AppLabelKey{AppID: currentApp.ID, OrgID: org.ID}).
		Find(&rows)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to get app labels: %w", res.Error))
		return
	}

	overrides := map[string]string{}
	if currentApp.LabelColors != nil {
		overrides = map[string]string(currentApp.LabelColors)
	}

	labels := make([]AppLabelKeySummary, 0, len(rows))
	for _, row := range rows {
		defaultColor := StandardLabelColors[row.ColorIndex%len(StandardLabelColors)]

		lk := AppLabelKeySummary{
			Key:          row.Key,
			DefaultColor: defaultColor,
			Values:       []string(row.Values),
			EntityTypes:  []string(row.EntityTypes),
			UsageCount:   row.UsageCount,
		}

		if color, ok := overrides[row.Key]; ok {
			lk.Color = color
			lk.IsOverride = true
		} else {
			lk.Color = defaultColor
		}

		labels = append(labels, lk)
	}

	ctx.JSON(http.StatusOK, AppLabelsResponse{
		Labels:        labels,
		LabelColors:   overrides,
		DefaultColors: StandardLabelColors,
	})
}
