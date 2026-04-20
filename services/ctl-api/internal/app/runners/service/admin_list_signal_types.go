package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals"
)

func (s *service) AdminListSignalTypes(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, signals.AllSignalTypes())
}
