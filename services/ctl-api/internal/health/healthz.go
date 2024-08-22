package health

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Service) GetHealthzHandler(ctx *gin.Context) {
	// NOTE: didn't find a ping command with gorm. Fetching first from org to check if DB is up
	org := app.Org{}
	res := s.db.WithContext(ctx).First(&org)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to query database: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
