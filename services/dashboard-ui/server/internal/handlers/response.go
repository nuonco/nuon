package handlers

import "github.com/gin-gonic/gin"

// respondJSON wraps data in the TAPIResponse shape expected by the frontend useQuery hook.
func respondJSON(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"data":    data,
		"error":   nil,
		"status":  status,
		"headers": gin.H{},
	})
}

// respondError wraps an error in the TAPIResponse shape expected by the frontend.
func respondError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{
		"data": nil,
		"error": gin.H{
			"error":       err.Error(),
			"description": err.Error(),
		},
		"status":  status,
		"headers": gin.H{},
	})
}
