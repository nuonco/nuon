package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type AccountHandler struct {
	l *zap.Logger
}

func NewAccountHandler(l *zap.Logger) *AccountHandler {
	return &AccountHandler{l: l}
}

func (h *AccountHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/account", h.GetAccount)
	return nil
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	account, err := client.GetCurrentUser(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusInternalServerError, err)
		return
	}

	respondJSON(c, http.StatusOK, account)
}
