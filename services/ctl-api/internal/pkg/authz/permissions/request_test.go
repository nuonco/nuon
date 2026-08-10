package permissions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestFromRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		method string
		want   Permission
	}{
		{method: http.MethodGet, want: PermissionRead},
		{method: http.MethodHead, want: PermissionRead},
		{method: http.MethodPost, want: PermissionCreate},
		{method: http.MethodPut, want: PermissionUpdate},
		{method: http.MethodPatch, want: PermissionUpdate},
		{method: http.MethodDelete, want: PermissionDelete},
		{method: http.MethodOptions, want: PermissionUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.method, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(tc.method, "/v1/apps", nil)
			assert.Equal(t, tc.want, FromRequest(ctx))
		})
	}
}
