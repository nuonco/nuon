package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *service) Index(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"title":       "Nuon Admin Dashboard",
		"description": "Welcome to the Nuon admin dashboard service",
	})
}
