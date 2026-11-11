package installcreate

import (
	"context"
	"testing"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type installAssignmentAPI struct {
	nuon.Client
	assignment        *models.ServiceInstallAssignmentResponse
	latestInputCalls  int
	appConfigCalls    int
	appConfigResponse *models.AppAppConfig
}

func (a *installAssignmentAPI) GetInstallAssignment(ctx context.Context, appID string, labels map[string]string, appBranchID string) (*models.ServiceInstallAssignmentResponse, error) {
	return a.assignment, nil
}

func (a *installAssignmentAPI) GetAppInputLatestConfig(context.Context, string) (*models.AppAppInputConfig, error) {
	a.latestInputCalls++
	return &models.AppAppInputConfig{ID: "latest-fallback"}, nil
}

func (a *installAssignmentAPI) GetAppConfig(context.Context, string, string, *bool) (*models.AppAppConfig, error) {
	a.appConfigCalls++
	return a.appConfigResponse, nil
}

func TestResolveInstallAssignmentWithNoMatchedBranchValidatesAgainstTheAssignmentsConfigNotLatest(t *testing.T) {
	api := &installAssignmentAPI{
		assignment: &models.ServiceInstallAssignmentResponse{AppConfigID: "fallback-cfg"},
		appConfigResponse: &models.AppAppConfig{
			ID:    "fallback-cfg",
			Input: &models.AppAppInputConfig{ID: "fallback-cfg-inputs"},
		},
	}

	assignment, inputCfg, err := ResolveInstallAssignment(context.Background(), api, "app", nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if assignment.AppConfigID != "fallback-cfg" {
		t.Fatalf("assignment = %#v", assignment)
	}
	if inputCfg == nil || inputCfg.ID != "fallback-cfg-inputs" {
		t.Fatalf("inputCfg = %#v", inputCfg)
	}
	if api.latestInputCalls != 0 {
		t.Fatalf("expected GetAppInputLatestConfig to not be called, got %d calls", api.latestInputCalls)
	}
}

func TestResolveInstallAssignmentWithMatchedBranchReturnsTheBranchAssignment(t *testing.T) {
	api := &installAssignmentAPI{
		assignment: &models.ServiceInstallAssignmentResponse{
			AppConfigID:   "branch-cfg",
			AppBranchID:   "branch-id",
			AppBranchName: "prod",
		},
		appConfigResponse: &models.AppAppConfig{
			ID:    "branch-cfg",
			Input: &models.AppAppInputConfig{ID: "branch-cfg-inputs"},
		},
	}

	assignment, inputCfg, err := ResolveInstallAssignment(context.Background(), api, "app", map[string]string{"env": "prod"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if assignment.AppBranchID != "branch-id" {
		t.Fatalf("assignment = %#v", assignment)
	}
	if inputCfg == nil || inputCfg.ID != "branch-cfg-inputs" {
		t.Fatalf("inputCfg = %#v", inputCfg)
	}
}
