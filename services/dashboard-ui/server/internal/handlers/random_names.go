package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-faker/faker/v4"
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
	name := fmt.Sprintf("%s-%s-%s",
		strings.ToLower(faker.Word()),
		strings.ToLower(faker.Word()),
		strings.ToLower(faker.Word()),
	)
	c.JSON(http.StatusOK, gin.H{"name": name})
}
