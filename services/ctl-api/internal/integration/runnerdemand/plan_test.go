package runnerdemand_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func testPlanCredentials(t *testing.T, s testServices, namespace string) {
	f := newFixture(t, s, namespace, 1, time.Minute)
	f.create(&app.VCSConnection{GithubAccountName: "Acme", GithubInstallID: "123"})
	const saved = `{"git_source":{"url":"https://x-access-token:expired@github.com/acme/api.git","ref":"0123456789abcdef0123456789abcdef01234567","path":"infra"},"vars":{"replicas":3},"future_field":{"preserved":true}}`
	require.NoError(t, s.DB.Model(&app.RunnerJobPlan{}).Where(app.RunnerJobPlan{RunnerJobID: f.jobs[0].ID}).Update("plan_json", saved).Error)
	f.start()
	f.awaitPark(0)
	require.NoError(t, s.DB.Scopes(scopes.WithDisableViews).Model(&f.jobs[0]).Updates(map[string]any{"group": app.RunnerJobGroupSandbox, "type": app.RunnerJobTypeSandboxTerraform}).Error)

	var denied atomic.Bool
	github := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/app/installations/123/access_tokens" {
			http.NotFound(w, r)
			return
		}
		var request struct {
			Repositories []string `json:"repositories"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case len(request.Repositories) == 0:
			_, _ = w.Write([]byte(`{"token":"installation-wide","repositories":[{"name":"other-repository"}]}`))
		case len(request.Repositories) != 1 || request.Repositories[0] != "api":
			http.Error(w, `{"message":"unexpected repository"}`, http.StatusUnprocessableEntity)
		case denied.Load():
			http.Error(w, `{"message":"repository access revoked"}`, http.StatusUnprocessableEntity)
		default:
			_, _ = w.Write([]byte(`{"token":"fresh-scoped-token","repositories":[{"name":"api"}]}`))
		}
	}))
	t.Cleanup(github.Close)
	oldBaseURL := s.GitHub.BaseURL
	s.GitHub.BaseURL, _ = url.Parse(github.URL + "/")
	t.Cleanup(func() { s.GitHub.BaseURL = oldBaseURL })

	var plan map[string]any
	f.request(http.MethodGet, f.jobURL(0)+"/plan", nil, http.StatusOK, &plan)
	var want map[string]any
	require.NoError(t, json.Unmarshal([]byte(saved), &want))
	want["git_source"].(map[string]any)["url"] = "https://x-access-token:fresh-scoped-token@github.com/acme/api.git"
	require.Equal(t, want, plan)
	var composite plantypes.CompositePlan
	f.request(http.MethodGet, f.jobURL(0)+"/composite-plan", nil, http.StatusOK, &composite)
	require.NotNil(t, composite.SandboxRunPlan)
	require.Equal(t, "https://x-access-token:fresh-scoped-token@github.com/acme/api.git", composite.SandboxRunPlan.GitSource.URL)
	require.Equal(t, "0123456789abcdef0123456789abcdef01234567", composite.SandboxRunPlan.GitSource.Ref)
	require.Equal(t, "infra", composite.SandboxRunPlan.GitSource.Path)
	require.EqualValues(t, 3, composite.SandboxRunPlan.Vars["replicas"])
	denied.Store(true)
	f.request(http.MethodGet, f.jobURL(0)+"/plan", nil, http.StatusInternalServerError, nil)
	f.request(http.MethodGet, f.jobURL(0)+"/composite-plan", nil, http.StatusInternalServerError, nil)
	var persisted app.RunnerJobPlan
	require.NoError(t, s.DB.Where(app.RunnerJobPlan{RunnerJobID: f.jobs[0].ID}).First(&persisted).Error)
	require.JSONEq(t, saved, persisted.PlanJSON)
}
