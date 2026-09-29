package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID VerifyCloudConnection
// @Summary verify a cloud connection
// @Tags cloud-connections
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Success 202 {object} ConnectionResponse
// @Router /v1/cloud-connections/{connection_id}/verify [post]
func (s *service) Verify(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	connection, err := s.verify(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	response, err := s.response(ctx, connection)
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusAccepted, response)
}

func (s *service) verify(ctx context.Context, orgID, connectionID string) (*app.CloudConnection, error) {
	connection, err := s.get(ctx, orgID, connectionID)
	if err != nil {
		return nil, err
	}
	now := s.now().UTC().Truncate(time.Microsecond)
	if verificationInProgress(connection) && connection.VerificationRequestedAt.After(now.Add(-2*time.Minute)) {
		return connection, nil
	}
	if err := s.enqueueVerification(ctx, connection); err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(&app.CloudConnection{}).Where(app.CloudConnection{OrgID: orgID, ID: connection.ID}).
		Select("verification_requested_at").Updates(app.CloudConnection{VerificationRequestedAt: &now}).Error; err != nil {
		return nil, fmt.Errorf("save cloud connection verification request: %w", err)
	}
	connection.VerificationRequestedAt = &now
	return connection, nil
}
