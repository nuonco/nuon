package service

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestListLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := map[string]struct {
		query string
		want  int
	}{
		"default":     {query: "", want: 50},
		"invalid":     {query: "?limit=nope", want: 50},
		"nonpositive": {query: "?limit=0", want: 50},
		"requested":   {query: "?limit=25", want: 25},
		"capped":      {query: "?limit=101", want: 100},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/"+tt.query, nil)
			assert.Equal(t, tt.want, listLimit(ctx))
		})
	}
}
