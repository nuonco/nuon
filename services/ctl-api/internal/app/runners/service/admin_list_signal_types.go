package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *service) AdminListSignalTypes(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, []string{})
}
