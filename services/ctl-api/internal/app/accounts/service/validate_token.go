package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						ValidateToken
// @Summary				Validate authentication token
// @Description			Returns 200 if the provided token is valid, 401 otherwise.
// @Tags					auth
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Success				200	{object}	map[string]bool
// @Failure				401	{object}	stderr.ErrResponse
// @Router					/v1/auth/validate [GET]
func (s *service) ValidateToken(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"valid": true})
}
