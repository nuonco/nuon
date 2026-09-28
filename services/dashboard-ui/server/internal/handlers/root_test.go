package handlers

import (
	"net/http"
	"testing"

	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/dashboard-ui/server/internal"
)

const (
	accountNoJourney    = `{"id":"acc1","user_journeys":[{"name":"evaluation","steps":[{"name":"a","complete":false}]}]}`
	accountFirstRunOpen = `{"id":"acc1","user_journeys":[{"name":"first_run","steps":[{"name":"start","complete":true},{"name":"connect","complete":false}]}]}`
	accountFirstRunDone = `{"id":"acc1","user_journeys":[{"name":"first_run","steps":[{"name":"start","complete":true},{"name":"connect","complete":true}]}]}`
	oneOrg              = `[{"id":"orgabc123","name":"acme"}]`
)

func TestRootHandle(t *testing.T) {
	tests := []struct {
		name      string
		orgCookie string
		orgs      string
		account   func(http.ResponseWriter, *http.Request)
		want      string
	}{
		{name: "new sign-up with no org", orgs: `[]`, want: "/onboarding"},
		{name: "mid-onboarding via org list", orgs: oneOrg, account: jsonResponse(http.StatusOK, accountFirstRunOpen), want: "/onboarding"},
		{name: "mid-onboarding via org cookie", orgCookie: "orgabc123", account: jsonResponse(http.StatusOK, accountFirstRunOpen), want: "/onboarding"},
		{name: "finished or skipped", orgs: oneOrg, account: jsonResponse(http.StatusOK, accountFirstRunDone), want: "/orgabc123"},
		{name: "finished via org cookie", orgCookie: "orgabc123", account: jsonResponse(http.StatusOK, accountFirstRunDone), want: "/orgabc123"},
		{name: "existing user without first_run", orgs: oneOrg, account: jsonResponse(http.StatusOK, accountNoJourney), want: "/orgabc123"},
		{name: "account lookup fails", orgCookie: "orgabc123", account: jsonResponse(http.StatusInternalServerError, `{"error":"boom"}`), want: "/orgabc123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCtlAPI{account: tt.account}
			if tt.orgs != "" {
				fake.orgs = jsonResponse(http.StatusOK, tt.orgs)
			}
			api := fake.start(t)
			h := NewRootHandler(&internal.Config{APIUrl: api.URL}, zap.NewNop())

			cookies := []*http.Cookie{authed()}
			if tt.orgCookie != "" {
				cookies = append(cookies, &http.Cookie{Name: orgCookie, Value: tt.orgCookie})
			}
			rec := serve(t, h.RegisterRoutes, "/", cookies...)

			if rec.Code != http.StatusFound {
				t.Fatalf("status = %d, want 302", rec.Code)
			}
			if got := rec.Header().Get("Location"); got != tt.want {
				t.Errorf("Location = %q, want %q", got, tt.want)
			}
		})
	}
}
