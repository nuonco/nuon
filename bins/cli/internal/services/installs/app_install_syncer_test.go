package installs

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/installcreate"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestInstallDiffKey(t *testing.T) {
	tests := []struct {
		name string
		key  string
		want string
	}{
		{
			name: "plain input passes through unchanged",
			key:  "sub_domain",
			want: "sub_domain",
		},
		{
			name: "helm values override decodes to components.<name>.helm_values",
			key:  config.HelmValuesOverrideInputName("whoami"),
			want: "components.whoami.helm_values",
		},
		{
			name: "tf vars override decodes to components.<name>.tf_vars",
			key:  config.TFVarsOverrideInputName("certificate"),
			want: "components.certificate.tf_vars",
		},
		{
			name: "component name with underscores/dashes round-trips",
			key:  config.HelmValuesOverrideInputName("foo-bar_baz"),
			want: "components.foo-bar_baz.helm_values",
		},
		{
			name: "non-override reserved-looking key passes through",
			key:  "inputs",
			want: "inputs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := installDiffKey(tt.key); got != tt.want {
				t.Fatalf("installDiffKey(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestInputDefaults(t *testing.T) {
	inputs := []*models.AppAppInput{
		{Name: "optional_with_default", Default: "opt"},
		{Name: "required_with_default", Required: true, Default: "req"},
		{Name: "required_no_default", Required: true},
		{Name: "sensitive_with_default", Required: true, Default: "secret", Sensitive: true},
		{Name: "customer_owned", Default: "cust", Source: string(models.AppAppInputSourceCustomer)},
	}

	got := inputDefaults(inputs)
	want := map[string]string{
		"optional_with_default": "opt",
		"required_with_default": "req",
	}

	if len(got) != len(want) {
		t.Fatalf("inputDefaults() = %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("inputDefaults()[%q] = %q, want %q", k, got[k], v)
		}
	}
}

type inputConfigAPI struct {
	nuon.Client
	appID, appConfigID string
	cfg                *models.AppAppConfig
}

func (a *inputConfigAPI) GetAppConfig(_ context.Context, appID, appConfigID string, _ *bool) (*models.AppAppConfig, error) {
	a.appID = appID
	a.appConfigID = appConfigID
	return a.cfg, nil
}

func TestInputConfigForNewInstallUsesLatestAppBranchRunConfig(t *testing.T) {
	inputCfg := &models.AppAppInputConfig{ID: "input-config-1"}
	api := &inputConfigAPI{
		cfg: &models.AppAppConfig{
			ID:    "app-config-2",
			Input: inputCfg,
		},
	}

	syncer := &appInstallSyncer{api: api, appID: "app-1"}
	got, err := syncer.inputConfigForNewInstall(context.Background(), &models.AppAppBranch{
		ID:   "branch-1",
		Name: "production",
		LatestRun: &models.AppAppBranchRun{
			AppConfigID: "app-config-2",
		},
	})
	if err != nil {
		t.Fatalf("expected latest branch config input: %v", err)
	}
	if got.ID != inputCfg.ID {
		t.Fatalf("input config ID = %q, want %q", got.ID, inputCfg.ID)
	}
	if api.appID != "app-1" {
		t.Fatalf("GetAppConfig appID = %q, want app-1", api.appID)
	}
	if api.appConfigID != "app-config-2" {
		t.Fatalf("GetAppConfig appConfigID = %q, want app-config-2", api.appConfigID)
	}
}

type branchConfigAPI struct {
	nuon.Client
	polls    []*models.AppAppBranch
	runs     []*models.AppWorkflow
	cfg      *models.AppAppConfig
	fetches  int
	runLists int
}

func (a *branchConfigAPI) GetAppBranch(context.Context, string, string) (*models.AppAppBranch, error) {
	a.fetches++
	if a.fetches >= len(a.polls) {
		return a.polls[len(a.polls)-1], nil
	}
	return a.polls[a.fetches], nil
}

func (a *branchConfigAPI) GetAppBranchRuns(context.Context, string, string) ([]*models.AppWorkflow, error) {
	a.runLists++
	return a.runs, nil
}

func (a *branchConfigAPI) GetAppConfig(context.Context, string, string, *bool) (*models.AppAppConfig, error) {
	return a.cfg, nil
}

func TestAwaitBranchConfig(t *testing.T) {
	origPoll, origTimeout := installcreate.BranchConfigPoll, installcreate.BranchConfigTimeout
	installcreate.BranchConfigPoll = time.Millisecond
	installcreate.BranchConfigTimeout = 50 * time.Millisecond
	t.Cleanup(func() {
		installcreate.BranchConfigPoll = origPoll
		installcreate.BranchConfigTimeout = origTimeout
	})

	t.Run("no app branch", func(t *testing.T) {
		got, err := (&appInstallSyncer{}).awaitBranchConfig(context.Background(), nil, false)
		if err != nil || got != nil {
			t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
		}
	})

	t.Run("already ready", func(t *testing.T) {
		api := &branchConfigAPI{cfg: &models.AppAppConfig{ID: "cfg", Status: models.AppAppConfigStatusActive}}
		branch := &models.AppAppBranch{ID: "b1", Name: "seed", LatestRun: &models.AppAppBranchRun{AppConfigID: "cfg"}}
		got, err := (&appInstallSyncer{api: api, appID: "app"}).awaitBranchConfig(context.Background(), branch, false)
		if err != nil {
			t.Fatal(err)
		}
		if got != branch || api.fetches != 0 {
			t.Fatalf("got branch %v after %d fetches", got, api.fetches)
		}
	})

	t.Run("waits while the run is in flight", func(t *testing.T) {
		ready := &models.AppAppBranch{ID: "b1", Name: "seed", LatestRun: &models.AppAppBranchRun{Status: "running", AppConfigID: "cfg"}}
		api := &branchConfigAPI{
			polls: []*models.AppAppBranch{nil, ready},
			cfg:   &models.AppAppConfig{ID: "cfg", Status: models.AppAppConfigStatusActive},
		}
		start := &models.AppAppBranch{ID: "b1", Name: "seed", LatestRun: &models.AppAppBranchRun{Status: "running"}}
		got, err := (&appInstallSyncer{api: api, appID: "app"}).awaitBranchConfig(context.Background(), start, false)
		if err != nil {
			t.Fatal(err)
		}
		if got != ready || api.fetches != 1 {
			t.Fatalf("got %v after %d fetches", got, api.fetches)
		}
	})

	t.Run("single failed run", func(t *testing.T) {
		api := &branchConfigAPI{runs: []*models.AppWorkflow{{ID: "run-1"}}}
		branch := &models.AppAppBranch{ID: "b1", Name: "seed", LatestRun: &models.AppAppBranchRun{Status: "failed"}}
		_, err := (&appInstallSyncer{api: api, appID: "app"}).awaitBranchConfig(context.Background(), branch, false)
		if err == nil || !strings.Contains(err.Error(), "run failed") {
			t.Fatalf("err = %v", err)
		}
		if api.fetches != 0 || api.runLists != 1 {
			t.Fatalf("fetches=%d runLists=%d", api.fetches, api.runLists)
		}
	})
}

func TestResolveRequiredInputs(t *testing.T) {
	tests := []struct {
		name     string
		inputs   []*models.AppAppInput
		defined  map[string]string
		current  map[string]string
		wantFill map[string]string
		wantErrs []string
	}{
		{
			name:     "required input with a default is not needed in the config",
			inputs:   []*models.AppAppInput{{Name: "runner_image_url", Required: true, Default: "upstream"}},
			defined:  map[string]string{},
			current:  map[string]string{},
			wantFill: map[string]string{"runner_image_url": "upstream"},
		},
		{
			name:     "required input with no default is still an error",
			inputs:   []*models.AppAppInput{{Name: "root_domain", Required: true}},
			defined:  map[string]string{},
			current:  map[string]string{},
			wantFill: map[string]string{},
			wantErrs: []string{"missing required input root_domain"},
		},
		{
			name:     "value already on the install is left alone, not reverted to the default",
			inputs:   []*models.AppAppInput{{Name: "runner_image_url", Required: true, Default: "upstream"}},
			defined:  map[string]string{},
			current:  map[string]string{"runner_image_url": "vendored"},
			wantFill: map[string]string{},
		},
		{
			name:     "config value wins over the default",
			inputs:   []*models.AppAppInput{{Name: "runner_image_url", Required: true, Default: "upstream"}},
			defined:  map[string]string{"runner_image_url": "explicit"},
			current:  map[string]string{},
			wantFill: map[string]string{},
		},
		{
			name: "customer owned input still cannot be set from the config",
			inputs: []*models.AppAppInput{
				{Name: "region", Required: true, Default: "us-west-2", Source: string(models.AppAppInputSourceCustomer)},
			},
			defined:  map[string]string{"region": "us-east-1"},
			current:  map[string]string{},
			wantFill: map[string]string{},
			wantErrs: []string{"refusing to set user_configurable input region"},
		},
		{
			name: "optional input with no value is not filled",
			inputs: []*models.AppAppInput{
				{Name: "sub_domain", Default: "app"},
			},
			defined:  map[string]string{},
			current:  map[string]string{},
			wantFill: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fill, errs := resolveRequiredInputs(tt.inputs, tt.defined, tt.current)

			if len(fill) != len(tt.wantFill) {
				t.Fatalf("fill = %v, want %v", fill, tt.wantFill)
			}
			for k, v := range tt.wantFill {
				if fill[k] != v {
					t.Errorf("fill[%q] = %q, want %q", k, fill[k], v)
				}
			}

			if len(errs) != len(tt.wantErrs) {
				t.Fatalf("errs = %v, want %v", errs, tt.wantErrs)
			}
			for i, want := range tt.wantErrs {
				if errs[i].Error() != want {
					t.Errorf("errs[%d] = %q, want %q", i, errs[i].Error(), want)
				}
			}
		})
	}
}

func TestUnknownInputsError(t *testing.T) {
	declaredOverride := config.HelmValuesOverrideInputName("whoami")
	tests := []struct {
		name    string
		inputs  []*models.AppAppInput
		defined map[string]string
		wantErr string
	}{
		{
			name:    "declared inputs are accepted",
			inputs:  []*models.AppAppInput{{Name: "region"}, {Name: declaredOverride}},
			defined: map[string]string{"region": "us-west-2", declaredOverride: "replicaCount: 1\n"},
		},
		{
			name:    "empty config is accepted",
			inputs:  []*models.AppAppInput{{Name: "region"}},
			defined: map[string]string{},
		},
		{
			name:    "name missing from the app config",
			inputs:  []*models.AppAppInput{{Name: "region"}},
			defined: map[string]string{"region": "us-west-2", "retired": "1"},
			wantErr: "unknown input(s): retired",
		},
		{
			name:    "undeclared component override uses the install config path",
			inputs:  []*models.AppAppInput{{Name: "region"}},
			defined: map[string]string{declaredOverride: "replicaCount: 1\n", "retired": "1"},
			wantErr: "unknown input(s): components.whoami.helm_values, retired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := unknownInputsError(tt.inputs, tt.defined)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("err = %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("err = %v, want %q", err, tt.wantErr)
			}
		})
	}
}

type dryRunInputsAPI struct {
	nuon.Client
	configs     map[string]*models.AppAppConfig
	latestInput *models.AppAppInputConfig
	fetched     []string
}

func (a *dryRunInputsAPI) GetAppConfig(_ context.Context, _, appConfigID string, _ *bool) (*models.AppAppConfig, error) {
	a.fetched = append(a.fetched, appConfigID)
	cfg, ok := a.configs[appConfigID]
	if !ok {
		return nil, fmt.Errorf("no app config %s", appConfigID)
	}
	return cfg, nil
}

func (a *dryRunInputsAPI) GetInstallCurrentInputs(context.Context, string) (*models.AppInstallInputs, error) {
	return &models.AppInstallInputs{Values: map[string]string{}}, nil
}

func (a *dryRunInputsAPI) GenerateCLIInstallConfig(context.Context, string) ([]byte, error) {
	return []byte("name = \"acme\"\n"), nil
}

func (a *dryRunInputsAPI) GetAppInputLatestConfig(context.Context, string) (*models.AppAppInputConfig, error) {
	return a.latestInput, nil
}

func appConfigWithInputs(id string, names ...string) *models.AppAppConfig {
	inputs := make([]*models.AppAppInput, len(names))
	for i, name := range names {
		inputs[i] = &models.AppAppInput{Name: name}
	}
	return &models.AppAppConfig{
		ID:     id,
		Status: models.AppAppConfigStatusActive,
		Input:  &models.AppAppInputConfig{Inputs: inputs},
	}
}

func TestSyncExistingInstallDryRunRejectsUnknownInput(t *testing.T) {
	api := &dryRunInputsAPI{
		configs: map[string]*models.AppAppConfig{
			"pin-config":    appConfigWithInputs("pin-config", "region"),
			"branch-config": appConfigWithInputs("branch-config", "region", "renamed"),
		},
	}
	install := &models.AppInstall{ID: "inst-1", Name: "acme", AppID: "app-1", AppConfigID: "pin-config"}
	branch := &models.AppAppBranch{
		ID:        "b1",
		Name:      "main",
		LatestRun: &models.AppAppBranchRun{AppConfigID: "branch-config"},
	}
	syncer := &appInstallSyncer{
		api:    api,
		appID:  "app-1",
		asJSON: true,
		branchesByInstall: map[string]*models.AppAppBranch{
			"acme": branch,
		},
	}
	cfg := &config.Install{
		Name: "acme",
		InputGroups: []config.InputGroup{{
			Inputs: map[string]string{"renamed": "1"},
		}},
	}

	_, err := syncer.syncExistingInstall(context.Background(), cfg, install, true, false, true)
	if err == nil || err.Error() != "unknown input(s): renamed" {
		t.Fatalf("err = %v", err)
	}
	sawPin := false
	for _, id := range api.fetched {
		if id == "pin-config" {
			sawPin = true
		}
	}
	if !sawPin {
		t.Fatalf("pinned app config was not loaded, fetched %v", api.fetched)
	}
}

func TestSyncExistingInstallDryRunAllowsInputDeclaredOnlyOnPin(t *testing.T) {
	api := &dryRunInputsAPI{
		configs: map[string]*models.AppAppConfig{
			"pin-config":    appConfigWithInputs("pin-config", "legacy"),
			"branch-config": appConfigWithInputs("branch-config", "region"),
		},
	}
	install := &models.AppInstall{ID: "inst-1", Name: "acme", AppID: "app-1", AppConfigID: "pin-config"}
	branch := &models.AppAppBranch{
		ID:        "b1",
		Name:      "main",
		LatestRun: &models.AppAppBranchRun{AppConfigID: "branch-config"},
	}
	syncer := &appInstallSyncer{
		api:    api,
		appID:  "app-1",
		asJSON: true,
		branchesByInstall: map[string]*models.AppAppBranch{
			"acme": branch,
		},
	}
	cfg := &config.Install{
		Name: "acme",
		InputGroups: []config.InputGroup{{
			Inputs: map[string]string{"legacy": "1"},
		}},
	}

	got, err := syncer.syncExistingInstall(context.Background(), cfg, install, true, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatalf("dry-run install = %#v, want nil", got)
	}
}

func TestSyncNewInstallDryRunRejectsUnknownInput(t *testing.T) {
	api := &dryRunInputsAPI{
		latestInput: &models.AppAppInputConfig{
			Inputs: []*models.AppAppInput{{Name: "region"}},
		},
	}
	syncer := &appInstallSyncer{api: api, appID: "app-1", asJSON: true}
	cfg := &config.Install{
		Name: "acme",
		InputGroups: []config.InputGroup{{
			Inputs: map[string]string{"retired": "1"},
		}},
		Components: map[string]config.ComponentOverride{
			"whoami": {HelmValues: "replicaCount: 1\n"},
		},
	}

	_, err := syncer.syncNewInstall(context.Background(), cfg, true, false, true)
	if err == nil || err.Error() != "unknown input(s): components.whoami.helm_values, retired" {
		t.Fatalf("err = %v", err)
	}
}
