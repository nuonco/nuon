package apps

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

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

func TestBranchSyncRunConfigDiff(t *testing.T) {
	repo := &config.ConnectedRepoConfig{Repo: "acme/platform", Branch: "main", Directory: "."}
	remoteLatest := func(run *models.AppAppBranchRunConfig) *models.AppAppBranchConfig {
		return &models.AppAppBranchConfig{
			ConnectedGithubVcsConfig: &models.AppConnectedGithubVCSConfig{Repo: repo.Repo, Branch: repo.Branch, Directory: repo.Directory},
			RunConfig:                run,
			InstallGroups:            []*models.AppAppBranchInstallGroup{{Name: defaultInstallGroupName, Default: true}},
		}
	}

	tests := []struct {
		name    string
		local   *config.AppBranchRunConfig
		remote  *models.AppAppBranchRunConfig
		changed bool
	}{
		{name: "absent matches push", local: nil, remote: &models.AppAppBranchRunConfig{Mode: models.AppAppBranchRunModePush}},
		{name: "absent matches remote without run config", local: nil, remote: nil},
		{name: "all alias matches push", local: &config.AppBranchRunConfig{Mode: "all"}, remote: &models.AppAppBranchRunConfig{Mode: models.AppAppBranchRunModePush}},
		{name: "on_tag_prefix alias matches on_tag", local: &config.AppBranchRunConfig{Mode: "on_tag_prefix", TagPrefix: "v"}, remote: &models.AppAppBranchRunConfig{Mode: models.AppAppBranchRunModeOnTag, TagPrefix: "v"}},
		{name: "mode change", local: &config.AppBranchRunConfig{Mode: "manual_only"}, remote: &models.AppAppBranchRunConfig{Mode: models.AppAppBranchRunModePush}, changed: true},
		{name: "mode removed", local: nil, remote: &models.AppAppBranchRunConfig{Mode: models.AppAppBranchRunModeManualOnly}, changed: true},
		{name: "tag prefix change", local: &config.AppBranchRunConfig{Mode: "on_tag", TagPrefix: "release-"}, remote: &models.AppAppBranchRunConfig{Mode: models.AppAppBranchRunModeOnTag, TagPrefix: "v"}, changed: true},
		{name: "github label change", local: &config.AppBranchRunConfig{Mode: "on_github_label", GithubLabel: "deploy"}, remote: &models.AppAppBranchRunConfig{Mode: models.AppAppBranchRunModeOnGithubLabel, GithubLabel: "ship"}, changed: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			resolver := testResolver(nil)
			local, err := canonicalizeLocalBranch(ctx, resolver, &config.AppBranchConfig{Name: "main", ConnectedRepo: repo, Run: tc.local})
			require.NoError(t, err)
			remote, err := normalizeRemoteBranch(ctx, resolver, "main", remoteLatest(tc.remote))
			require.NoError(t, err)

			plan, err := buildBranchSyncPlan(
				[]*config.AppBranchConfig{local},
				[]*models.AppAppBranch{{ID: "br-main", Name: "main", ManagedBy: appBranchManagedByConfig}},
				map[string]*config.AppBranchConfig{"main": remote},
				true,
			)
			require.NoError(t, err)
			require.Len(t, plan, 1)
			want := branchOpUnchanged
			if tc.changed {
				want = branchOpUpdate
			}
			require.Equal(t, want, plan[0].Op)
		})
	}
}

func TestBranchSyncRunConfigRepoLessBranchUnchanged(t *testing.T) {
	ctx := context.Background()
	resolver := testResolver(nil)
	local, err := canonicalizeLocalBranch(ctx, resolver, &config.AppBranchConfig{Name: "qa"})
	require.NoError(t, err)
	remote, err := normalizeRemoteBranch(ctx, resolver, "qa", nil)
	require.NoError(t, err)
	require.False(t, local.Diff(remote).Summary().HasChanged)
}
