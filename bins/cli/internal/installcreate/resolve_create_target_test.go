package installcreate

import (
	"context"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type createTargetAPI struct {
	nuon.Client
	target            *models.ServiceInstallCreateTargetResponse
	latestInputCalls  int
	appConfigCalls    int
	appConfigResponse *models.AppAppConfig
}

func (a *createTargetAPI) GetInstallCreateTarget(ctx context.Context, appID string, labels map[string]string, appBranchID string) (*models.ServiceInstallCreateTargetResponse, error) {
	return a.target, nil
}

func (a *createTargetAPI) GetAppInputLatestConfig(context.Context, string) (*models.AppAppInputConfig, error) {
	a.latestInputCalls++
	return &models.AppAppInputConfig{ID: "latest-fallback"}, nil
}

func (a *createTargetAPI) GetAppConfig(context.Context, string, string, *bool) (*models.AppAppConfig, error) {
	a.appConfigCalls++
	return a.appConfigResponse, nil
}

func TestResolveCreateTargetWithNoMatchedBranchValidatesAgainstTheTargetsConfigNotLatest(t *testing.T) {
	api := &createTargetAPI{
		target: &models.ServiceInstallCreateTargetResponse{AppConfigID: "fallback-cfg"},
		appConfigResponse: &models.AppAppConfig{
			ID:    "fallback-cfg",
			Input: &models.AppAppInputConfig{ID: "fallback-cfg-inputs"},
		},
	}

	target, inputCfg, err := ResolveCreateTarget(context.Background(), api, "app", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if target.AppConfigID != "fallback-cfg" {
		t.Fatalf("target = %#v", target)
	}
	if inputCfg == nil || inputCfg.ID != "fallback-cfg-inputs" {
		t.Fatalf("inputCfg = %#v", inputCfg)
	}
	if api.latestInputCalls != 0 {
		t.Fatalf("expected GetAppInputLatestConfig to not be called, got %d calls", api.latestInputCalls)
	}
}

func TestResolveCreateTargetWithMatchedBranchReturnsTheBranchTarget(t *testing.T) {
	api := &createTargetAPI{
		target: &models.ServiceInstallCreateTargetResponse{
			AppConfigID:   "branch-cfg",
			AppBranchID:   "branch-id",
			AppBranchName: "prod",
		},
		appConfigResponse: &models.AppAppConfig{
			ID:    "branch-cfg",
			Input: &models.AppAppInputConfig{ID: "branch-cfg-inputs"},
		},
	}

	target, inputCfg, err := ResolveCreateTarget(context.Background(), api, "app", map[string]string{"env": "prod"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if target.AppBranchID != "branch-id" {
		t.Fatalf("target = %#v", target)
	}
	if inputCfg == nil || inputCfg.ID != "branch-cfg-inputs" {
		t.Fatalf("inputCfg = %#v", inputCfg)
	}
}
