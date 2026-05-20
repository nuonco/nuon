package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *service) FullSweep(c *gin.Context) {
	resp, err := s.enqueuer.FullSweep(c.Request.Context())
	if err != nil {
		s.l.Error("full sweep failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to run full sweep"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
