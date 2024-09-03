package health

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"go.temporal.io/sdk/client"
)

func (s *Service) GetLivezHandler(ctx *gin.Context) {
	// NOTE: didn't find a ping command with gorm. For now will try fetching from org.
	org := app.Org{}
	dbRes := s.db.WithContext(ctx).First(&org)
	if dbRes.Error != nil {
		ctx.Error(fmt.Errorf("unable to query database: %w", dbRes.Error))
		return
	}

	_, err := s.tclient.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check temporal health: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})

}
