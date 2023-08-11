package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Service) GetVersionHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"git_ref": s.gitRef,
	})
}
