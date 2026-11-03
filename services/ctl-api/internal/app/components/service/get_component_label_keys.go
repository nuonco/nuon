package service

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
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
func (s *service) GetComponentLabelKeys(ctx *gin.Context) (any, error) {
	access, err := require.ReadAccessFromContext(ctx.Request.Context())
	if err != nil {
		return nil, err
	}
	components, err := access.Query()
	if err != nil {
		return nil, err
	}

	appID := ctx.Param("app_id")

	type kv struct {
		Key   string
		Value string
	}
	var rows []kv
	query := access.DB().WithContext(ctx.Request.Context()).
		Raw(`SELECT DISTINCT key, value FROM (?) AS visible_components,
			LATERAL jsonb_each_text(visible_components.labels) ORDER BY key, value`,
			components.Select("components.labels").Where(&app.Component{AppID: appID}))

	if err := query.Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to get component labels: %w", err)
	}

	result := make(map[string][]string)
	for _, r := range rows {
		result[r.Key] = append(result[r.Key], r.Value)
	}

	return result, nil
}
