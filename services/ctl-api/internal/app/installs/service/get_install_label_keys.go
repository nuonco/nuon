package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID						GetInstallLabelKeys
// @Summary				get distinct label key:value pairs across all installs for an org
// @Description			Returns all distinct label key:value pairs for installs in the current org.
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Success				200	{object}	map[string][]string
// @Router					/v1/installs/label-keys [GET]
func (s *service) GetInstallLabelKeys(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	result, err := s.getDistinctLabels(ctx, "installs", "org_id = ? AND deleted_at = 0", org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// getDistinctLabels queries a table for all distinct label key→values.
// Returns map[string][]string where keys are label keys and values are sorted distinct values.
func (s *service) getDistinctLabels(ctx *gin.Context, table, where string, args ...any) (map[string][]string, error) {
	type kv struct {
		Key   string
		Value string
	}

	var rows []kv
	query := s.db.WithContext(ctx).
		Raw("SELECT DISTINCT key, value FROM (SELECT (jsonb_each_text(labels)).* FROM "+table+" WHERE labels IS NOT NULL AND labels != '{}'::jsonb AND "+where+") AS t ORDER BY key, value", args...)

	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, r := range rows {
		result[r.Key] = append(result[r.Key], r.Value)
	}

	return result, nil
}
