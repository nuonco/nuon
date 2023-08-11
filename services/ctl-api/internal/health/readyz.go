package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) GetReadyzHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
