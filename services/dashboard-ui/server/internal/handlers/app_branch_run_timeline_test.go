package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	nuon "github.com/nuonco/nuon/sdks/nuon-go"
)

func branchRunsQueryFor(t *testing.T, url string) *nuon.GetAppBranchRunsQuery {
	t.Helper()
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, url, nil)

	query := appBranchRunsQuery(c)
	if query.Planonly == nil {
		t.Fatal("planonly was nil")
	}
	return query
}

func TestAppBranchRunsQueryDefaults(t *testing.T) {
	query := branchRunsQueryFor(t, "/sse")

	if !*query.Planonly {
		t.Error("planonly should default to true so previews are included")
	}
	if query.Limit != 10 || query.Offset != 0 {
		t.Errorf("limit/offset = %d/%d, want 10/0", query.Limit, query.Offset)
	}
	if query.Type != "" || query.Status != "" || query.Q != "" ||
		query.CreatedAtGte != "" || query.CreatedAtLte != "" {
		t.Errorf("filters should be empty by default, got %+v", query)
	}
}

func TestAppBranchRunsQueryCarriesEveryRESTFilter(t *testing.T) {
	query := branchRunsQueryFor(t, "/sse?planonly=false"+
		"&type=app_branches_manual_update,app_branch_config_update"+
		"&status=error,cancelled&q=pr+42"+
		"&created_at_gte=2026-09-01T00:00:00Z&created_at_lte=2026-09-10T00:00:00Z"+
		"&limit=25&offset=50")

	if *query.Planonly {
		t.Error("planonly=false should exclude previews")
	}
	if want := "app_branches_manual_update,app_branch_config_update"; query.Type != want {
		t.Errorf("type = %q, want %q", query.Type, want)
	}
	if want := "error,cancelled"; query.Status != want {
		t.Errorf("status = %q, want %q", query.Status, want)
	}
	if want := "pr 42"; query.Q != want {
		t.Errorf("q = %q, want %q", query.Q, want)
	}
	if want := "2026-09-01T00:00:00Z"; query.CreatedAtGte != want {
		t.Errorf("created_at_gte = %q, want %q", query.CreatedAtGte, want)
	}
	if want := "2026-09-10T00:00:00Z"; query.CreatedAtLte != want {
		t.Errorf("created_at_lte = %q, want %q", query.CreatedAtLte, want)
	}
	if query.Limit != 25 || query.Offset != 50 {
		t.Errorf("limit/offset = %d/%d, want 25/50", query.Limit, query.Offset)
	}
}
