package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *service) DeleteOrg(ctx *gin.Context) {
	orgID := ctx.Param("id")

	err := s.deleteOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusAccepted, map[string]string{
		"status": "ok",
	})
}

func (s *service) deleteOrg(ctx context.Context, orgID string) error {
	res := s.db.WithContext(ctx).Delete(&app.Org{
		Model: app.Model{
			ID: orgID,
		},
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}

	return nil
}
