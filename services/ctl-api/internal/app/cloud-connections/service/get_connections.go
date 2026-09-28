package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

// @ID ListCloudConnections
// @Summary list cloud connections
// @Param offset query int false "offset of results to return" Default(0)
// @Param limit query int false "limit of results to return" Default(10)
// @Param page query int false "page number of results to return" Default(0)
// @Param q query string false "search by name, account, or role ARN"
// @Tags cloud-connections
// @Produce json
// @Security APIKey
// @Security OrgID
// @Success 200 {array} ConnectionListResponse
// @Router /v1/cloud-connections [get]
func (s *service) List(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var connections []app.CloudConnection
	query := s.db.WithContext(ctx).Scopes(scopes.WithOffsetPagination).Where(app.CloudConnection{OrgID: org.ID})
	if q := ctx.Query("q"); q != "" {
		pattern := "%" + q + "%"
		query = query.Where("name ILIKE ? OR target_id ILIKE ? OR principal ILIKE ?", pattern, pattern, pattern)
	}
	if err := query.Order("created_at DESC, id DESC").Find(&connections).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to list cloud connections: %w", err))
		return
	}
	connections, err = db.HandlePaginatedResponse(ctx, connections)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to handle paginated response: %w", err))
		return
	}
	ids := make([]string, 0, len(connections))
	for i := range connections {
		ids = append(ids, connections[i].ID)
	}
	usage, err := s.usages(ctx, ids)
	if err != nil {
		ctx.Error(err)
		return
	}
	responses := make([]ConnectionListResponse, 0, len(connections))
	for i := range connections {
		responses = append(responses, newListResponse(&connections[i], usage))
	}
	ctx.JSON(http.StatusOK, responses)
}
