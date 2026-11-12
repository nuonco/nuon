package installcreate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type latestInputAPI struct {
	nuon.Client
	calls int
}

func (a *latestInputAPI) GetAppInstallInputConfig(context.Context, string, string) (*models.AppAppInputConfig, error) {
	a.calls++
	return &models.AppAppInputConfig{ID: "input"}, nil
}

type pinnedInputAPI struct {
	nuon.Client
	latestCalls int
	configCalls int
	cfg         *models.AppAppConfig
}

func (a *pinnedInputAPI) GetAppInputLatestConfig(context.Context, string) (*models.AppAppInputConfig, error) {
	a.latestCalls++
	return &models.AppAppInputConfig{ID: "latest"}, nil
}

func (a *pinnedInputAPI) GetAppConfig(context.Context, string, string, *bool) (*models.AppAppConfig, error) {
	a.configCalls++
	return a.cfg, nil
}

func TestInputConfigForInstallUsesPinnedConfig(t *testing.T) {
	api := &pinnedInputAPI{cfg: &models.AppAppConfig{
		Input: &models.AppAppInputConfig{
			ID: "branch-inputs",
			Inputs: []*models.AppAppInput{
				{Name: "branch_only", GroupID: "g1"},
			},
			InputGroups: []*models.AppAppInputGroup{{ID: "g1"}},
		},
	}}
	got, err := InputConfigForInstall(context.Background(), api, &models.AppInstall{
		AppID:       "app",
		AppConfigID: "cfg-branch",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "branch-inputs" || api.latestCalls != 0 || api.configCalls != 1 {
		t.Fatalf("got %s latest=%d config=%d", got.ID, api.latestCalls, api.configCalls)
	}
	if len(got.InputGroups) != 1 || len(got.InputGroups[0].AppInputs) != 1 || got.InputGroups[0].AppInputs[0].Name != "branch_only" {
		t.Fatalf("groups = %#v", got.InputGroups)
	}
}

func TestInputConfigForInstallWithoutPinUsesLatest(t *testing.T) {
	api := &pinnedInputAPI{}
	got, err := InputConfigForInstall(context.Background(), api, &models.AppInstall{AppID: "app"})
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "latest" || api.latestCalls != 1 || api.configCalls != 0 {
		t.Fatalf("got %s latest=%d config=%d", got.ID, api.latestCalls, api.configCalls)
	}
}

func TestResolveInputConfigWithoutBranchDoesNotPoll(t *testing.T) {
	api := &latestInputAPI{}
	got, err := ResolveInputConfig(context.Background(), api, "app", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "input" || api.calls != 1 {
		t.Fatalf("got %v after %d calls", got, api.calls)
	}
}

func inputConfigServer(t *testing.T, installInputConfig http.HandlerFunc) (nuon.Client, *int) {
	t.Helper()
	latestCalls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/apps/app/install-input-config", installInputConfig)
	mux.HandleFunc("/v1/apps/app/input-latest-config", func(w http.ResponseWriter, _ *http.Request) {
		latestCalls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"legacy-latest"}`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	api, err := nuon.New(nuon.WithURL(srv.URL), nuon.WithAuthToken("token"), nuon.WithOrgID("org"))
	if err != nil {
		t.Fatal(err)
	}
	return api, &latestCalls
}

func TestResolveInputConfigWithoutBranchFallsBackOnMissingRoute(t *testing.T) {
	api, latestCalls := inputConfigServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 page not found"))
	})

	got, err := ResolveInputConfig(context.Background(), api, "app", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "legacy-latest" || *latestCalls != 1 {
		t.Fatalf("got %s latestCalls=%d", got.ID, *latestCalls)
	}
}

func TestResolveInputConfigWithoutBranchReturnsServerNotFound(t *testing.T) {
	api, latestCalls := inputConfigServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"app not found","description":"App not found.","user_error":true}`))
	})

	_, err := ResolveInputConfig(context.Background(), api, "app", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if *latestCalls != 0 {
		t.Fatalf("latestCalls=%d", *latestCalls)
	}
}

func TestResolveInputConfigWithoutBranchReturnsUserError(t *testing.T) {
	api, latestCalls := inputConfigServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"no config","description":"Choose an app branch for this install.","user_error":true}`))
	})

	_, err := ResolveInputConfig(context.Background(), api, "app", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if *latestCalls != 0 {
		t.Fatalf("latestCalls=%d", *latestCalls)
	}
}
