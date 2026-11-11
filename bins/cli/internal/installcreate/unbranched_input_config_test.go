package installcreate

import (
	"context"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type unbranchedConfigAPI struct {
	nuon.Client
	pages        [][]*models.AppAppConfig
	configs      map[string]*models.AppAppConfig
	branches     []*models.AppAppBranch
	configsCalls int
	getCalls     int
}

func (a *unbranchedConfigAPI) GetAppConfigs(_ context.Context, _ string, _ *models.GetPaginatedQuery) ([]*models.AppAppConfig, bool, error) {
	page := a.configsCalls
	a.configsCalls++
	if page >= len(a.pages) {
		return nil, false, nil
	}
	return a.pages[page], page < len(a.pages)-1, nil
}

func (a *unbranchedConfigAPI) GetAppConfig(_ context.Context, _ string, configID string, _ *bool) (*models.AppAppConfig, error) {
	a.getCalls++
	return a.configs[configID], nil
}

func (a *unbranchedConfigAPI) GetAppBranches(_ context.Context, _ string, _ *models.GetPaginatedQuery) ([]*models.AppAppBranch, bool, error) {
	return a.branches, false, nil
}

func TestUnbranchedInputConfigSkipsBranchConfigs(t *testing.T) {
	api := &unbranchedConfigAPI{
		pages: [][]*models.AppAppConfig{{
			{ID: "branch-cfg", AppBranchID: "branch-1", Status: models.AppAppConfigStatusActive},
			{ID: "unbranched-cfg", Status: models.AppAppConfigStatusActive},
		}},
		configs: map[string]*models.AppAppConfig{
			"unbranched-cfg": {ID: "unbranched-cfg", Input: &models.AppAppInputConfig{ID: "unbranched-inputs"}},
		},
	}

	got, err := UnbranchedInputConfig(context.Background(), api, "app")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "unbranched-inputs" {
		t.Fatalf("got %s", got.ID)
	}
}

func TestUnbranchedInputConfigSkipsGitPreviewConfigs(t *testing.T) {
	api := &unbranchedConfigAPI{
		pages: [][]*models.AppAppConfig{{
			{
				ID:     "preview-cfg",
				Status: models.AppAppConfigStatusActive,
				Labels: models.GithubComNuoncoNuonPkgLabelsLabels{"source": string(models.AppAppBranchRunTypeGitDashPreviewDashRun)},
			},
			{ID: "good-cfg", Status: models.AppAppConfigStatusActive},
		}},
		configs: map[string]*models.AppAppConfig{
			"good-cfg": {ID: "good-cfg", Input: &models.AppAppInputConfig{ID: "good-inputs"}},
		},
	}

	got, err := UnbranchedInputConfig(context.Background(), api, "app")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "good-inputs" {
		t.Fatalf("got %s", got.ID)
	}
}

func TestUnbranchedInputConfigSkipsInactiveConfigs(t *testing.T) {
	api := &unbranchedConfigAPI{
		pages: [][]*models.AppAppConfig{{
			{ID: "pending-cfg", Status: models.AppAppConfigStatusPending},
			{ID: "active-cfg", Status: models.AppAppConfigStatusActive},
		}},
		configs: map[string]*models.AppAppConfig{
			"active-cfg": {ID: "active-cfg", Input: &models.AppAppInputConfig{ID: "active-inputs"}},
		},
	}

	got, err := UnbranchedInputConfig(context.Background(), api, "app")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "active-inputs" {
		t.Fatalf("got %s", got.ID)
	}
}

func TestUnbranchedInputConfigFindsMatchOnLaterPage(t *testing.T) {
	api := &unbranchedConfigAPI{
		pages: [][]*models.AppAppConfig{
			{{ID: "branch-cfg", AppBranchID: "branch-1", Status: models.AppAppConfigStatusActive}},
			{{ID: "unbranched-cfg", Status: models.AppAppConfigStatusActive}},
		},
		configs: map[string]*models.AppAppConfig{
			"unbranched-cfg": {ID: "unbranched-cfg", Input: &models.AppAppInputConfig{ID: "unbranched-inputs"}},
		},
	}

	got, err := UnbranchedInputConfig(context.Background(), api, "app")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "unbranched-inputs" {
		t.Fatalf("got %s", got.ID)
	}
	if api.configsCalls != 2 {
		t.Fatalf("expected 2 pages fetched, got %d", api.configsCalls)
	}
}

func TestUnbranchedInputConfigNoneFoundWithoutBranches(t *testing.T) {
	api := &unbranchedConfigAPI{
		pages: [][]*models.AppAppConfig{{}},
	}

	_, err := UnbranchedInputConfig(context.Background(), api, "app")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "no active app config from nuon apps sync; run nuon apps sync first" {
		t.Fatalf("got %q", err.Error())
	}
}

func TestUnbranchedInputConfigNoneFoundWithBranches(t *testing.T) {
	api := &unbranchedConfigAPI{
		pages:    [][]*models.AppAppConfig{{}},
		branches: []*models.AppAppBranch{{ID: "branch-1"}},
	}

	_, err := UnbranchedInputConfig(context.Background(), api, "app")
	if err == nil {
		t.Fatal("expected error")
	}
	want := "no active app config from nuon apps sync; this app's configs are on app branches; set app_branch to create the install on an app branch"
	if err.Error() != want {
		t.Fatalf("got %q want %q", err.Error(), want)
	}
}
