package apps

import (
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestBuildBranchSyncPlan_CreateUpdateDeleteUnchanged(t *testing.T) {
	local := []*config.AppBranchConfig{
		{Name: "main", ConnectedRepo: &config.ConnectedRepoConfig{Repo: "acme/platform", Branch: "main", Directory: "."}},
		{Name: "staging", ConnectedRepo: &config.ConnectedRepoConfig{Repo: "acme/platform", Branch: "staging", Directory: "."}},
		{Name: "qa"},
	}
	remotes := []*models.AppAppBranch{
		{ID: "br-main", Name: "main", ManagedBy: appBranchManagedByConfig},
		{ID: "br-qa", Name: "qa", ManagedBy: appBranchManagedByConfig},
		{ID: "br-legacy", Name: "legacy", ManagedBy: appBranchManagedByConfig},
		{ID: "br-manual-only", Name: "scratch", ManagedBy: appBranchManagedByManually},
	}
	remoteCfg := map[string]*config.AppBranchConfig{
		"main": {
			Name:          "main",
			ConnectedRepo: &config.ConnectedRepoConfig{Repo: "acme/platform", Branch: "old", Directory: "."},
		},
		"qa":     {Name: "qa"},
		"legacy": {Name: "legacy"},
	}

	plan, err := buildBranchSyncPlan(local, remotes, remoteCfg, true)
	if err != nil {
		t.Fatal(err)
	}

	got := map[string]branchOp{}
	for _, item := range plan {
		got[item.Name] = item.Op
	}
	if got["staging"] != branchOpCreate {
		t.Fatalf("staging op = %s", got["staging"])
	}
	if got["main"] != branchOpUpdate {
		t.Fatalf("main op = %s", got["main"])
	}
	if got["qa"] != branchOpUnchanged {
		t.Fatalf("qa op = %s", got["qa"])
	}
	if got["legacy"] != branchOpDelete {
		t.Fatalf("legacy op = %s", got["legacy"])
	}
	if _, ok := got["scratch"]; ok {
		t.Fatal("manually managed absent branch should not be deleted")
	}
}

func TestBuildBranchSyncPlan_SingleFileDoesNotPrune(t *testing.T) {
	local := []*config.AppBranchConfig{{Name: "main"}}
	remotes := []*models.AppAppBranch{
		{ID: "br-main", Name: "main", ManagedBy: appBranchManagedByConfig},
		{ID: "br-other", Name: "other", ManagedBy: appBranchManagedByConfig},
	}
	remoteCfg := map[string]*config.AppBranchConfig{
		"main":  {Name: "main"},
		"other": {Name: "other"},
	}

	plan, err := buildBranchSyncPlan(local, remotes, remoteCfg, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range plan {
		if item.Op == branchOpDelete {
			t.Fatalf("single-file mode deleted %s", item.Name)
		}
	}
}

func TestBuildBranchSyncPlan_ManualSameNameErrors(t *testing.T) {
	local := []*config.AppBranchConfig{{Name: "main"}}
	remotes := []*models.AppAppBranch{
		{ID: "br-main", Name: "main", ManagedBy: appBranchManagedByManually},
	}

	_, err := buildBranchSyncPlan(local, remotes, map[string]*config.AppBranchConfig{"main": {Name: "main"}}, true)
	if err == nil || !strings.Contains(err.Error(), "managed manually") {
		t.Fatalf("expected manual-branch error, got %v", err)
	}
}

func TestAppBranchConfigDiff_CoversInstallGroupAutoApprove(t *testing.T) {
	old := &config.AppBranchConfig{
		Name: "main",
		InstallGroups: []config.AppBranchInstallGroupConfig{{
			Name:                         "canary",
			AutoApproveOnPoliciesPassing: generics.ToPtr(false),
		}},
	}
	newCfg := &config.AppBranchConfig{
		Name: "main",
		InstallGroups: []config.AppBranchInstallGroupConfig{{
			Name:                         "canary",
			AutoApproveOnPoliciesPassing: generics.ToPtr(true),
		}},
	}
	d := newCfg.Diff(old)
	if !d.Summary().HasChanged {
		t.Fatal("expected auto-approve change")
	}
}
