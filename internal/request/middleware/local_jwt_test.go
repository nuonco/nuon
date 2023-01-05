package middleware

import (
	"net/http"
	"testing"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"github.com/stretchr/testify/assert"
)

func TestLocalJWTMiddleware(t *testing.T) {
	tests := []struct {
		description string
		status      int
		cb          gin.HandlerFunc
	}{
		{
			description: "should set a token on the context",
			status:      http.StatusOK,
			cb: func(c *gin.Context) {
				val := c.Request.Context().Value(jwtmiddleware.ContextKey{})
				assert.NotNil(t, val)
			},
		},

		{
			description: "should set the correct email on the context",
			status:      http.StatusOK,
			cb: func(c *gin.Context) {
				ctx := c.Request.Context()
				claim, err := parseJWT(ctx)
				assert.Nil(t, err)
				assert.NotNil(t, claim)
				assert.Equal(t, claim.ExternalID, "test@nuon.co")
			},
		},
	}

	// execute tests
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			router := gin.New()
			router.HandleMethodNotAllowed = true

			middleware := NewLocalJWTMiddleware("test@nuon.co")
			router.Use(adapter.Wrap(middleware.InjectJWT))
			router.Use(test.cb)

			PerformRequest(router, "GET", "/")
		})
	}
}
