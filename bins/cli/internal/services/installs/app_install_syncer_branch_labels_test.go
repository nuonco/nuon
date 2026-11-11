package installs

import (
	"context"
	"fmt"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type labelTargetAPI struct {
	nuon.Client
	target      *models.ServiceInstallCreateTargetResponse
	targetErr   error
	branches    map[string]*models.AppAppBranch
	configs     map[string]*models.AppAppConfig
	targetCalls []map[string]string
	latestCalls int
	created     []*models.ServiceCreateInstallRequest
}

func (a *labelTargetAPI) GetInstallCreateTarget(_ context.Context, _ string, lbls map[string]string, _ string) (*models.ServiceInstallCreateTargetResponse, error) {
	a.targetCalls = append(a.targetCalls, lbls)
	return a.target, a.targetErr
}

func (a *labelTargetAPI) GetAppBranch(_ context.Context, _, branchID string) (*models.AppAppBranch, error) {
	branch, ok := a.branches[branchID]
	if !ok {
		return nil, fmt.Errorf("no app branch %s", branchID)
	}
	return branch, nil
}

func (a *labelTargetAPI) GetAppConfig(_ context.Context, _, appConfigID string, _ *bool) (*models.AppAppConfig, error) {
	cfg, ok := a.configs[appConfigID]
	if !ok {
		return nil, fmt.Errorf("no app config %s", appConfigID)
	}
	return cfg, nil
}

func (a *labelTargetAPI) GetAppInputLatestConfig(context.Context, string) (*models.AppAppInputConfig, error) {
	a.latestCalls++
	return &models.AppAppInputConfig{}, nil
}

func (a *labelTargetAPI) CreateInstall(_ context.Context, _ string, req *models.ServiceCreateInstallRequest) (*models.AppInstall, error) {
	a.created = append(a.created, req)
	return &models.AppInstall{ID: "inst-new", Name: *req.Name, WorkflowID: "wf-1"}, nil
}

func (a *labelTargetAPI) GetWorkflow(context.Context, string) (*models.AppWorkflow, error) {
	return nil, fmt.Errorf("no workflow")
}

func branchMatchedByLabelsAPI() *labelTargetAPI {
	return &labelTargetAPI{
		target: &models.ServiceInstallCreateTargetResponse{
			AppConfigID:    "branch-config",
			AppBranchID:    "branch-internal",
			AppBranchName:  "acme-internal",
			AppBranchGroup: "all",
		},
		branches: map[string]*models.AppAppBranch{
			"branch-internal": {
				ID:        "branch-internal",
				Name:      "acme-internal",
				LatestRun: &models.AppAppBranchRun{AppConfigID: "branch-config"},
			},
		},
		configs: map[string]*models.AppAppConfig{
			"branch-config": appConfigWithInputs("branch-config", "region", "executor_memory"),
		},
	}
}

func labeledInstallConfig() *config.Install {
	return &config.Install{
		Name:   "acme",
		Labels: map[string]string{"deploy_group": "acme-internal"},
		InputGroups: []config.InputGroup{{
			Inputs: map[string]string{"executor_memory": "4Gi"},
		}},
	}
}

func newLabelSyncer(api nuon.Client) *appInstallSyncer {
	return &appInstallSyncer{
		api:               api,
		appID:             "app-1",
		asJSON:            true,
		branchesByInstall: map[string]*models.AppAppBranch{},
	}
}

func TestSyncNewInstallMatchedByLabelsValidatesAgainstTheBranchConfig(t *testing.T) {
	api := branchMatchedByLabelsAPI()
	syncer := newLabelSyncer(api)

	_, err := syncer.syncNewInstall(context.Background(), labeledInstallConfig(), true, false, false)
	if err != nil {
		t.Fatalf("expected the branch-only input to be accepted, got: %v", err)
	}
	if len(api.targetCalls) != 1 || api.targetCalls[0]["deploy_group"] != "acme-internal" {
		t.Fatalf("expected one create-target lookup with the install's labels, got %v", api.targetCalls)
	}
	if api.latestCalls != 0 {
		t.Fatalf("expected GetAppInputLatestConfig to not be called, got %d calls", api.latestCalls)
	}
	if len(api.created) != 1 {
		t.Fatalf("expected one create request, got %d", len(api.created))
	}
	if api.created[0].AppBranchID != "" {
		t.Fatalf("expected the server to choose the branch, request sent AppBranchID %q", api.created[0].AppBranchID)
	}
	if got := syncer.branchesByInstall["acme"]; got == nil || got.ID != "branch-internal" {
		t.Fatalf("expected the matched branch to be recorded for the install, got %#v", got)
	}
}

func TestSyncNewInstallMatchedByLabelsDryRunDoesNotCreate(t *testing.T) {
	api := branchMatchedByLabelsAPI()
	syncer := newLabelSyncer(api)

	got, err := syncer.syncNewInstall(context.Background(), labeledInstallConfig(), true, false, true)
	if err != nil {
		t.Fatalf("expected dry run to accept the branch-only input, got: %v", err)
	}
	if got != nil {
		t.Fatalf("dry-run install = %#v, want nil", got)
	}
	if len(api.targetCalls) != 1 {
		t.Fatalf("expected one create-target lookup, got %d", len(api.targetCalls))
	}
	if len(api.created) != 0 {
		t.Fatalf("expected no create request on dry run, got %d", len(api.created))
	}
}

func TestSyncNewInstallWithoutLabelsValidatesAgainstTheFallbackTarget(t *testing.T) {
	api := &labelTargetAPI{
		target: &models.ServiceInstallCreateTargetResponse{AppConfigID: "fallback-config"},
		configs: map[string]*models.AppAppConfig{
			"fallback-config": appConfigWithInputs("fallback-config", "region"),
		},
	}
	syncer := newLabelSyncer(api)
	cfg := &config.Install{
		Name: "acme",
		InputGroups: []config.InputGroup{{
			Inputs: map[string]string{"region": "us-west-2"},
		}},
	}

	_, err := syncer.syncNewInstall(context.Background(), cfg, true, false, false)
	if err != nil {
		t.Fatalf("expected the fallback config's input to be accepted, got: %v", err)
	}
	if api.latestCalls != 0 {
		t.Fatalf("expected GetAppInputLatestConfig to not be called, got %d calls", api.latestCalls)
	}
	if len(api.created) != 1 || api.created[0].AppBranchID != "" {
		t.Fatalf("expected one create request without a branch, got %#v", api.created)
	}
	if got := syncer.branchesByInstall["acme"]; got != nil {
		t.Fatalf("expected no branch recorded for the install, got %#v", got)
	}
}

func TestSyncNewInstallSurfacesCreateTargetUserErrorsBeforeCreating(t *testing.T) {
	for name, description := range map[string]string{
		"labels match two branches":                              "The install's labels match more than one app branch (acme-a, acme-b). Set app_branch on the install to choose one.",
		"labels match no branch and there is no fallback config": "This install has no app_branch and its labels don't match any app branch's install groups. Add labels that a branch selects, or set app_branch.",
	} {
		t.Run(name, func(t *testing.T) {
			api := &labelTargetAPI{
				targetErr: &operations.GetInstallCreateTargetBadRequest{Payload: &models.StderrErrResponse{
					UserError:   true,
					Error:       "invalid request",
					Description: description,
				}},
			}
			syncer := newLabelSyncer(api)

			_, err := syncer.syncNewInstall(context.Background(), labeledInstallConfig(), true, false, false)
			if err == nil {
				t.Fatal("expected an error")
			}
			userErr, ok := nuon.ToUserError(err)
			if !ok || userErr.Description != description {
				t.Fatalf("expected the server's user error to be surfaced, got: %v", err)
			}
			if len(api.created) != 0 {
				t.Fatalf("expected no create request, got %d", len(api.created))
			}
		})
	}
}

func TestSyncNewInstallWithExplicitAppBranchSkipsLabelMatching(t *testing.T) {
	api := branchMatchedByLabelsAPI()
	api.branches["branch-explicit"] = &models.AppAppBranch{
		ID:        "branch-explicit",
		Name:      "acme-explicit",
		LatestRun: &models.AppAppBranchRun{AppConfigID: "explicit-config"},
	}
	api.configs["explicit-config"] = appConfigWithInputs("explicit-config", "region", "executor_memory")
	syncer := newLabelSyncer(api)
	syncer.branchesByInstall["acme"] = api.branches["branch-explicit"]

	_, err := syncer.syncNewInstall(context.Background(), labeledInstallConfig(), true, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(api.targetCalls) != 0 {
		t.Fatalf("expected no create-target lookup with an explicit app_branch, got %d", len(api.targetCalls))
	}
	if len(api.created) != 1 || api.created[0].AppBranchID != "branch-explicit" {
		t.Fatalf("expected one create request on the explicit branch, got %#v", api.created)
	}
}
