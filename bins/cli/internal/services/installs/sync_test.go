package installs

import (
	"context"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// TestParseInstallConfig_RejectsMalformedOverride proves the CLI parse path
// validates component-override syntax (Parse + Validate) before any API call,
// so a malformed Helm values / tfvars override fails fast at config load.
func TestParseInstallConfig_RejectsMalformedOverride(t *testing.T) {
	t.Run("invalid helm_values", func(t *testing.T) {
		raw := `name = "abcd"

[components.whoami]
helm_values = """
replicaCount: [5
"""
`
		_, err := parseInstallConfig(strings.NewReader(raw))
		if err == nil {
			t.Fatal("expected error for malformed helm_values, got nil")
		}
		if !strings.Contains(err.Error(), "helm_values") {
			t.Fatalf("error should mention helm_values, got: %v", err)
		}
	})

	t.Run("invalid tf_vars", func(t *testing.T) {
		raw := `name = "abcd"

[components.vpc]
tf_vars = """
cidr = =
"""
`
		_, err := parseInstallConfig(strings.NewReader(raw))
		if err == nil {
			t.Fatal("expected error for malformed tf_vars, got nil")
		}
		if !strings.Contains(err.Error(), "tf_vars") {
			t.Fatalf("error should mention tf_vars, got: %v", err)
		}
	})

	t.Run("valid overrides parse", func(t *testing.T) {
		raw := `name = "abcd"

[components.whoami]
helm_values = """
replicaCount: 5
"""

[components.vpc]
tf_vars = """
cidr = "10.0.0.0/16"
"""
`
		cfg, err := parseInstallConfig(strings.NewReader(raw))
		if err != nil {
			t.Fatalf("expected valid config to parse, got: %v", err)
		}
		if len(cfg.Components) != 2 {
			t.Fatalf("expected 2 component overrides, got %d", len(cfg.Components))
		}
	})
}

type testAppBranchLister struct {
	nuon.Client
	branches []*models.AppAppBranch
}

func (l testAppBranchLister) GetAppBranches(context.Context, string, *models.GetPaginatedQuery) ([]*models.AppAppBranch, bool, error) {
	return l.branches, false, nil
}

func TestResolveInstallConfigBranchesRequiresBranchWhenAppSyncDisabled(t *testing.T) {
	_, err := resolveInstallConfigBranches(
		context.Background(),
		testAppBranchLister{},
		"app-1",
		[]*config.Install{{Name: "production"}},
		true,
	)
	if err == nil {
		t.Fatal("expected app_branch requirement error")
	}
	if !strings.Contains(err.Error(), "must set app_branch when disable-app-sync is enabled") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResolveInstallConfigBranchesAcceptsNameAndID(t *testing.T) {
	branches := []*models.AppAppBranch{
		{ID: "branch-1", Name: "production"},
		{ID: "branch-2", Name: "staging"},
	}
	configs := []*config.Install{
		{Name: "one", AppBranch: "production"},
		{Name: "two", AppBranch: "branch-2"},
	}

	resolved, err := resolveInstallConfigBranches(
		context.Background(),
		testAppBranchLister{branches: branches},
		"app-1",
		configs,
		true,
	)
	if err != nil {
		t.Fatalf("expected branches to resolve: %v", err)
	}
	if resolved["one"].ID != "branch-1" || resolved["two"].ID != "branch-2" {
		t.Fatalf("unexpected resolved branches: %#v", resolved)
	}
	if configs[0].AppBranch != "production" || configs[1].AppBranch != "staging" {
		t.Fatalf("expected canonical branch names, got %q and %q", configs[0].AppBranch, configs[1].AppBranch)
	}
}
