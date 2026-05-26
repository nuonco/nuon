package replica

import (
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
)

func OptIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(routing.WithReplica(c.Request.Context()))
		c.Next()
	}
}

func OptOut() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(routing.WithoutReplica(c.Request.Context()))
		c.Next()
	}
}
