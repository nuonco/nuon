package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type AdminOrgDetails struct {
	*app.Org
}

// @ID			AdminListOrgsDetails
// @BasePath	/v1/orgs
// @Summary		Compact admin list of orgs with status
// @Description	Admin list of orgs intended for status / README rollups. The optional `status` query parameter filters by `status_v2->>'status'` and may be repeated.
// @Tags		orgs/admin
// @Security	AdminEmail
// @Accept		json
// @Produce		json
// @Param		offset	query	int			false	"offset of results to return"				Default(0)
// @Param		limit	query	int			false	"limit of results to return"				Default(10)
// @Param		page	query	int			false	"page number of results to return"			Default(0)
// @Param		status	query	[]string	false	"filter by composite status (repeatable)"	collectionFormat(multi)
// @Success		200	{array}	AdminOrgDetails
// @Router		/v1/orgs/details [GET]
func (s *service) AdminListOrgsDetails(ctx *gin.Context) {
	statuses := ctx.QueryArray("status")

	orgs, err := s.listOrgsDetails(ctx, statuses)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) listOrgsDetails(ctx *gin.Context, statuses []string) ([]*AdminOrgDetails, error) {
	var orgs []*app.Org
	tx := s.db.WithContext(ctx).
		Scopes(scopes.WithOffsetPagination).
		Order("created_at desc")
	if len(statuses) > 0 {
		tx = tx.Where("status_v2->>'status' IN ?", statuses)
	}
	if err := tx.Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("unable to list org details: %w", err)
	}

	orgs, err := db.HandlePaginatedResponse(ctx, orgs)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	items := make([]*AdminOrgDetails, 0, len(orgs))
	for _, o := range orgs {
		items = append(items, &AdminOrgDetails{Org: o})
	}

	return items, nil
}
