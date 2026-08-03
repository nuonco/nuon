package permissions

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestObject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const orgID = "org_one"

	tests := []struct {
		name   string
		method string
		path   string
		want   string
	}{
		{name: "app component build", method: http.MethodPost, path: "/v1/apps/app_one/components/cmp_one/builds", want: ComponentBuildsObject(orgID)},
		{name: "component build", method: http.MethodPost, path: "/v1/components/cmp_one/builds", want: ComponentBuildsObject(orgID)},
		{name: "build cancel", method: http.MethodPost, path: "/v1/apps/app_one/components/cmp_one/builds/bld_one/cancel", want: orgID},
		{name: "other post", method: http.MethodPost, path: "/v1/apps", want: orgID},
		{name: "read build route", method: http.MethodGet, path: "/v1/components/cmp_one/builds", want: orgID},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			router := gin.New()
			router.Handle(tc.method, routePattern(tc.path), func(ctx *gin.Context) {
				got = RequestObject(ctx, orgID)
			})
			req := httptest.NewRequest(tc.method, tc.path, nil)
			router.ServeHTTP(httptest.NewRecorder(), req)
			assert.Equal(t, tc.want, got)
		})
	}

	assert.NotEqual(t, ComponentBuildsObject("org_two"), ComponentBuildsObject(orgID))
}

func routePattern(path string) string {
	switch path {
	case "/v1/apps/app_one/components/cmp_one/builds":
		return "/v1/apps/:app_id/components/:component_id/builds"
	case "/v1/components/cmp_one/builds":
		return "/v1/components/:component_id/builds"
	case "/v1/apps/app_one/components/cmp_one/builds/bld_one/cancel":
		return "/v1/apps/:app_id/components/:component_id/builds/:build_id/cancel"
	default:
		return path
	}
}
