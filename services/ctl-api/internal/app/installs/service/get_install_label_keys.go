package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
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

	result := make(map[string][]string)
	err = authz.New(authz.Params{DB: s.db}).ReadCustomerResources(ctx.Request.Context(), org.ID, func(tx, _, installs *gorm.DB) error {
		var rows []struct{ Key, Value string }
		if err := tx.Raw("SELECT DISTINCT key, value FROM (?) AS permitted, LATERAL jsonb_each_text(permitted.labels) ORDER BY key, value", installs.Select("labels")).Scan(&rows).Error; err != nil {
			return err
		}
		for _, row := range rows {
			result[row.Key] = append(result[row.Key], row.Value)
		}
		return nil
	})
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}
