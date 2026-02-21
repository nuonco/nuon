package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
	"github.com/nuonco/nuon/services/dashboard-ui/server/internal/pkg/cctx"
)

type AccountHandler struct {
	cfg *internal.Config
	l   *zap.Logger
}

func NewAccountHandler(cfg *internal.Config, l *zap.Logger) *AccountHandler {
	return &AccountHandler{cfg: cfg, l: l}
}

func (h *AccountHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/account", h.GetAccount)
	return nil
}

func (h *AccountHandler) GetAccount(c *gin.Context) {
	// Auth middleware has already validated token and called GetCurrentUser
	// User data is available via apiclient middleware
	client, err := cctx.APIClientFromGinContext(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"data":  nil,
			"error": gin.H{"error": "internal error", "description": "API client not available"},
		})
		return
	}

	me, err := client.GetCurrentUser(c.Request.Context())
	if err != nil {
		h.l.Error("failed to get current user", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{
			"data":  nil,
			"error": gin.H{"error": "unauthorized", "description": "failed to get user"},
		})
		return
	}

	// Transform to account response format expected by frontend
	account := gin.H{
		"id":            me.ID,
		"email":         me.Email,
		"name":          me.Email, // Use email as name if no name available
		"org_ids":       me.OrgIds,
		"user_journeys": me.UserJourneys,
		"created_at":    me.CreatedAt,
		"updated_at":    me.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   account,
		"error":  nil,
		"status": http.StatusOK,
	})
}
