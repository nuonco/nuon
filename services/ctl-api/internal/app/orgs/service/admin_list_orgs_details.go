package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type AdminOrgDetails struct {
	*app.Org
}

const adminOrgDetailsDefaultLimit = 25

// @ID			AdminListOrgsDetails
// @BasePath	/v1/orgs
// @Summary	Return a compact admin list of orgs with their status
// @Description	Admin list of orgs intended for status / README rollups.
// @Description	Pagination is uncapped on this admin endpoint — pass any `limit`.
// @Description	The optional `status` query parameter filters by
// @Description	`status_v2->>'status'` and may be repeated.
// @Param			offset	query	int			false	"offset of results to return"	Default(0)
// @Param			limit	query	int			false	"limit of results to return (no upper cap)"	Default(25)
// @Param			status	query	[]string	false	"filter by composite status (repeatable)"	collectionFormat(multi)
// @Tags			orgs/admin
// @Security		AdminEmail
// @Accept			json
// @Produce		json
// @Success		200	{array}	AdminOrgDetails
// @Router			/v1/orgs/details [GET]
func (s *service) AdminListOrgsDetails(ctx *gin.Context) {
	limit, offset := parseAdminDetailsPagination(ctx)
	statuses := ctx.QueryArray("status")

	orgs, err := s.listOrgsDetails(ctx, limit, offset, statuses)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) listOrgsDetails(ctx *gin.Context, limit, offset int, statuses []string) ([]*AdminOrgDetails, error) {
	var orgs []*app.Org
	tx := s.db.WithContext(ctx).
		Order("created_at desc").
		Limit(limit).
		Offset(offset)
	if len(statuses) > 0 {
		tx = tx.Where("status_v2->>'status' IN ?", statuses)
	}
	if err := tx.Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("unable to list org details: %w", err)
	}

	items := make([]*AdminOrgDetails, 0, len(orgs))
	for _, o := range orgs {
		items = append(items, &AdminOrgDetails{Org: o})
	}

	return items, nil
}

func parseAdminDetailsPagination(ctx *gin.Context) (limit, offset int) {
	limit = adminOrgDetailsDefaultLimit
	if v, err := strconv.Atoi(ctx.Query("limit")); err == nil && v > 0 {
		limit = v
	}
	if v, err := strconv.Atoi(ctx.Query("offset")); err == nil && v >= 0 {
		offset = v
	}
	return limit, offset
}
