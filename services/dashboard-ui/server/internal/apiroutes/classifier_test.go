package apiroutes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func newTestClassifier(templates []string) *Classifier {
	c := NewClassifier("http://example.test", zap.NewNop())
	c.LoadTemplates(templates)
	return c
}

func TestMatch(t *testing.T) {
	templates := []string{
		"/v1/apps",
		"/v1/apps/{app_id}",
		"/v1/apps/{app_id}/installs",
		"/v1/apps/{app_id}/installs/{install_id}",
		"/v1/apps/current",
		"/v1/orgs/{org_id}/members/{member_id}",
	}
	c := newTestClassifier(templates)

	cases := []struct {
		name string
		path string
		want string
	}{
		{"collection", "/v1/apps", "/v1/apps"},
		{"single param", "/v1/apps/app98e2wpzdxwoey393edtqj45", "/v1/apps/{app_id}"},
		{"nested param", "/v1/apps/app98e2wpzdxwoey393edtqj45/installs/inl98e2wpzdxwoey393edtqj45", "/v1/apps/{app_id}/installs/{install_id}"},
		{"literal preferred over param", "/v1/apps/current", "/v1/apps/current"},
		{"two params", "/v1/orgs/org123/members/mem456", "/v1/orgs/{org_id}/members/{member_id}"},
		{"trailing slash tolerated", "/v1/apps/app98e2wpzdxwoey393edtqj45/", "/v1/apps/{app_id}"},
		{"unknown route bucketed", "/v1/wp-admin", UnmatchedEndpoint},
		{"random scan path bucketed", "/v1/aaaa/bbbb/cccc", UnmatchedEndpoint},
		{"partial path bucketed", "/v1", UnmatchedEndpoint},
		{"deeper than known bucketed", "/v1/apps/app123/installs/inl456/extra", UnmatchedEndpoint},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := c.Match(tc.path); got != tc.want {
				t.Errorf("Match(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestMatchBeforeLoad(t *testing.T) {
	c := NewClassifier("http://example.test", zap.NewNop())
	if got := c.Match("/v1/apps/app123"); got != UnmatchedEndpoint {
		t.Errorf("Match before spec load = %q, want %q", got, UnmatchedEndpoint)
	}
}

func TestRefresh(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/oapi/v3" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`{"paths":{"/v1/apps/{app_id}":{"get":{}},"/v1/orgs":{"get":{}}}}`))
	}))
	defer srv.Close()

	c := NewClassifier(srv.URL, zap.NewNop())
	if err := c.refresh(t.Context()); err != nil {
		t.Fatalf("refresh: %v", err)
	}

	if got := c.Match("/v1/apps/app98e2wpzdxwoey393edtqj45"); got != "/v1/apps/{app_id}" {
		t.Errorf("after refresh Match = %q, want /v1/apps/{app_id}", got)
	}
	if got := c.Match("/v1/orgs"); got != "/v1/orgs" {
		t.Errorf("after refresh Match = %q, want /v1/orgs", got)
	}
}
