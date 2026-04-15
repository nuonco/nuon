package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) AdminGetSandboxTemplates(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, app.DefaultSandboxTemplates())
}
