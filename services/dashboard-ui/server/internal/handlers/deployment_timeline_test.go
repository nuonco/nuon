package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func deploymentTimelineQueryFor(t *testing.T, url string) (limit, offset int, values map[string]string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, url, nil)

	query := deploymentTimelineQuery(c)
	return query.Limit, query.Offset, map[string]string{
		"type":           query.Type,
		"status":         query.Status,
		"resource":       query.Resource,
		"search":         query.Search,
		"created_at_gte": query.CreatedAtGte,
		"created_at_lte": query.CreatedAtLte,
	}
}

func TestDeploymentTimelineQueryDefaults(t *testing.T) {
	limit, offset, values := deploymentTimelineQueryFor(t, "/sse")
	if limit != 10 || offset != 0 {
		t.Errorf("limit/offset = %d/%d, want 10/0", limit, offset)
	}
	for key, value := range values {
		if value != "" {
			t.Errorf("%s = %q, want empty", key, value)
		}
	}
}

func TestDeploymentTimelineQueryCarriesEveryRESTFilter(t *testing.T) {
	limit, offset, values := deploymentTimelineQueryFor(t, "/sse?"+
		"type=provision,component_deploy"+
		"&status=success,error"+
		"&resource=payments"+
		"&search=manual+deploy"+
		"&created_at_gte=2026-09-01T00:00:00Z"+
		"&created_at_lte=2026-09-10T00:00:00Z"+
		"&limit=20&offset=40")

	if limit != 20 || offset != 40 {
		t.Errorf("limit/offset = %d/%d, want 20/40", limit, offset)
	}
	want := map[string]string{
		"type":           "provision,component_deploy",
		"status":         "success,error",
		"resource":       "payments",
		"search":         "manual deploy",
		"created_at_gte": "2026-09-01T00:00:00Z",
		"created_at_lte": "2026-09-10T00:00:00Z",
	}
	for key, value := range want {
		if values[key] != value {
			t.Errorf("%s = %q, want %q", key, values[key], value)
		}
	}
}
