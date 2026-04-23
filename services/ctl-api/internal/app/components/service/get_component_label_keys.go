package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetComponentLabelKeys
// @Summary				get distinct label key:value pairs across all components for an app
// @Description			Returns all distinct label key:value pairs for components in the given app.
// @Param					app_id	path	string	true	"app ID"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Success				200	{object}	map[string][]string
// @Router					/v1/apps/{app_id}/components/label-keys [GET]
func (s *service) GetComponentLabelKeys(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	type kv struct {
		Key   string
		Value string
	}
	var rows []kv
	query := s.db.WithContext(ctx).
		Raw(`SELECT DISTINCT key, value FROM (
			SELECT (jsonb_each_text(labels)).* FROM components
			WHERE labels IS NOT NULL AND labels != '{}'::jsonb
			AND app_id = ? AND deleted_at = 0
			AND EXISTS (SELECT 1 FROM apps WHERE apps.id = components.app_id AND apps.org_id = ?)
		) AS t ORDER BY key, value`, appID, org.ID)

	if err := query.Scan(&rows).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get component labels: %w", err))
		return
	}

	result := make(map[string][]string)
	for _, r := range rows {
		result[r.Key] = append(result[r.Key], r.Value)
	}

	ctx.JSON(http.StatusOK, result)
}
