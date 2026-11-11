package installcreate

import (
	"context"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

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

func TestResolveInputConfigWithoutBranchUsesTheCreateTargetConfig(t *testing.T) {
	api := &createTargetAPI{
		target: &models.ServiceInstallCreateTargetResponse{AppConfigID: "fallback-cfg"},
		appConfigResponse: &models.AppAppConfig{
			ID:     "fallback-cfg",
			Status: models.AppAppConfigStatusActive,
			Input:  &models.AppAppInputConfig{ID: "fallback-cfg-inputs"},
		},
	}
	got, err := ResolveInputConfig(context.Background(), api, "app", "")
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "fallback-cfg-inputs" {
		t.Fatalf("got %v", got)
	}
	if api.latestInputCalls != 0 {
		t.Fatalf("expected GetAppInputLatestConfig to not be called, got %d calls", api.latestInputCalls)
	}
}
