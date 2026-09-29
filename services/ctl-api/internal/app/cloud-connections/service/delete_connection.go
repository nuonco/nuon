package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// @ID DeleteCloudConnection
// @Summary delete a cloud connection
// @Tags cloud-connections
// @Security APIKey
// @Security OrgID
// @Param connection_id path string true "connection ID"
// @Success 204
// @Failure 409 {object} stderr.ErrResponse
// @Router /v1/cloud-connections/{connection_id} [delete]
func (s *service) Delete(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	err = s.delete(ctx, org.ID, ctx.Param("connection_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *service) delete(ctx context.Context, orgID, connectionID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var connection app.CloudConnection
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.CloudConnection{OrgID: orgID, ID: connectionID}).First(&connection).Error; err != nil {
			return fmt.Errorf("cloud connection not found: %w", err)
		}
		var installReferences int64
		if err := tx.Model(&app.Install{}).Where(app.Install{CloudConnectionID: &connection.ID}).Count(&installReferences).Error; err != nil {
			return fmt.Errorf("count install references: %w", err)
		}
		if installReferences > 0 {
			description := fmt.Sprintf("This connection is used by %d installs. Delete those installs first.", installReferences)
			if installReferences == 1 {
				description = "This connection is used by 1 install. Delete that install first."
			}
			return stderr.ErrConflict{Err: errors.New(description), Description: description}
		}
		if err := s.helpers.TerminateConnectionQueue(ctx, connection.ID); err != nil {
			return fmt.Errorf("terminate cloud connection queue: %w", err)
		}
		return tx.Delete(&connection).Error
	})
}
