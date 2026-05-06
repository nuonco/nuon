package service

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (s *service) Promote(c *gin.Context) {
	tag := s.cfg.RunnerContainerImageTag

	body, err := json.Marshal(map[string]string{"tag": tag})
	if err != nil {
		s.l.Error("failed to marshal promotion request", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
		return
	}

	s.proxyToInternalAPI(c, "POST", "/v1/general/promotion", bytes.NewReader(body))
}
