package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *service) Livez(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
