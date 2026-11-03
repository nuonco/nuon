package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
)

type OrgListItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// @ID                     GetOrgs
// @Summary                Return current user's orgs
// @Description.markdown   get_orgs.md
// @Param                  q                           query   string  false   "search query to filter orgs by name or ID"
// @Param                  offset                      query   int     false   "offset of results to return"   Default(0)
// @Param                  limit                       query   int     false   "limit of results to return"    Default(10)
// @Param                  page                        query   int     false   "page number of results to return"   Default(0)
// @Tags                   orgs
// @Accept                 json
// @Produce                json
// @Security               APIKey
// @Failure                400 {object} stderr.ErrResponse
// @Failure                401 {object} stderr.ErrResponse
// @Failure                403 {object} stderr.ErrResponse
// @Failure                404 {object} stderr.ErrResponse
// @Failure                500 {object} stderr.ErrResponse
// @Success                200 {array}  OrgListItem
// @Router                 /v1/orgs [GET]
func (s *service) GetCurrentUserOrgs(ctx *gin.Context) {
	limit, offset := 10, 0
	if page := cctx.OffsetPaginationFromContext(ctx); page != nil {
		limit, offset = page.Limit, page.Offset
	}
	orgs, err := s.getCustomerOrgs(ctx.Request.Context(), ctx.Query("q"), limit, offset)
	if err != nil {
		ctx.Error(err)
		return
	}
	orgs, err = db.HandlePaginatedResponse(ctx, orgs)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, orgs)
}

func (s *service) getCustomerOrgs(ctx context.Context, q string, limit, offset int) ([]OrgListItem, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, stderr.NewInvalidRequest(fmt.Errorf("invalid org pagination"))
	}
	orgs := []OrgListItem{}
	err := authz.New(authz.Params{DB: s.db}).ReadCustomerNavigation(ctx, func(tx *gorm.DB, orgIDs []string) error {
		query := tx.Model(&app.Org{}).Select("id", "name").Where("id IN ?", orgIDs).
			Order("LOWER(name), id").Limit(limit + 1).Offset(offset)
		if q = strings.TrimSpace(q); q != "" {
			pattern := "%" + strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(q) + "%"
			query = query.Where("name ILIKE ? OR id ILIKE ?", pattern, pattern)
		}
		return query.Find(&orgs).Error
	})
	return orgs, err
}
