package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID GetCloudConnection
// @Summary get a cloud connection
// @Tags cloud-connections
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Success 200 {object} ConnectionResponse
// @Router /v1/cloud-connections/{connection_id} [get]
func (s *service) Get(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.get(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, connection)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func (s *service) get(ctx context.Context, orgID, id string) (*app.CloudConnection, error) {
	var connection app.CloudConnection
	result := s.db.WithContext(ctx).Where(app.CloudConnection{OrgID: orgID, ID: id}).First(&connection)
	if result.Error != nil {
		return nil, fmt.Errorf("cloud connection not found: %w", result.Error)
	}
	return &connection, nil
}
