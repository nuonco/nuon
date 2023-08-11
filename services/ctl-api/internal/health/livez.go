package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) GetLivezHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
