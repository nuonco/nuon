package replica

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/routing"
)

type middleware struct{}

func (m middleware) Name() string {
	return "replica"
}

// Handler marks GET requests so that read queries are routed to the replica
// database via the routing.ConnPool.
func (m middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Request = c.Request.WithContext(routing.WithReplica(c.Request.Context()))
		}
		c.Next()
	}
}

func New() *middleware {
	return &middleware{}
}
