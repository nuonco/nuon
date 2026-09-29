package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/pkg/generics"
)

type RandomNamesHandler struct{}

func NewRandomNamesHandler() *RandomNamesHandler {
	return &RandomNamesHandler{}
}

func (h *RandomNamesHandler) RegisterRoutes(e *gin.Engine) error {
	e.GET("/api/random-name", h.RandomName)
	return nil
}

func (h *RandomNamesHandler) RandomName(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"name": generics.DefaultName()})
}
