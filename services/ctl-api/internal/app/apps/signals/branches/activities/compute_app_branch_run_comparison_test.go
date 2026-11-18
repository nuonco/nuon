package activities

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRunCommitSHA(t *testing.T) {
	sha := "abc123"
	require.Equal(t, sha, runCommitSHA(&app.AppBranchRun{
		VCSConnectionCommit: &app.VCSConnectionCommit{SHA: sha},
		HeadSHA:             "ignored",
	}))
	require.Equal(t, "head", runCommitSHA(&app.AppBranchRun{HeadSHA: "head"}))
	require.Equal(t, "", runCommitSHA(&app.AppBranchRun{}))
}

func connectedBranchConfig(id, repo, branch string) app.AppBranchConfig {
	return app.AppBranchConfig{
		ID:                       id,
		ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{ID: id + "-vcs", Repo: repo, Branch: branch},
	}
}

func TestRunBranchConfig(t *testing.T) {
	latest := connectedBranchConfig("latest", "acme/new", "main")
	pinned := connectedBranchConfig("pinned", "acme/old", "main")

	for name, tt := range map[string]struct {
		run    *app.AppBranchRun
		branch *app.AppBranch
		wantID string
	}{
		"uses the run's own config": {
			run:    &app.AppBranchRun{AppBranchConfig: pinned},
			branch: &app.AppBranch{Configs: []app.AppBranchConfig{latest}},
			wantID: "pinned",
		},
		"uses the run's public git config": {
			run: &app.AppBranchRun{AppBranchConfig: app.AppBranchConfig{
				ID:                 "public",
				PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "https://github.com/acme/public.git"},
			}},
			branch: &app.AppBranch{Configs: []app.AppBranchConfig{latest}},
			wantID: "public",
		},
		"falls back to the branch's latest config when the run's has no vcs config": {
			run:    &app.AppBranchRun{AppBranchConfig: app.AppBranchConfig{ID: "unloaded"}},
			branch: &app.AppBranch{Configs: []app.AppBranchConfig{latest}},
			wantID: "latest",
		},
		"falls back to the branch when there is no run": {
			branch: &app.AppBranch{Configs: []app.AppBranchConfig{latest}},
			wantID: "latest",
		},
		"nil when neither has a config": {
			run:    &app.AppBranchRun{},
			branch: &app.AppBranch{},
		},
		"nil when the branch is nil": {
			run: &app.AppBranchRun{},
		},
	} {
		t.Run(name, func(t *testing.T) {
			got := runBranchConfig(tt.run, tt.branch)
			if tt.wantID == "" {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, tt.wantID, got.ID)
		})
	}
}

func TestBranchRepoChanged(t *testing.T) {
	for name, tt := range map[string]struct {
		base app.AppBranchConfig
		head *app.AppBranchConfig
		want bool
	}{
		"same repo": {
			base: connectedBranchConfig("base", "acme/app", "main"),
			head: ptrConfig(connectedBranchConfig("head", "acme/app", "main")),
		},
		"same repo on a different git branch": {
			base: connectedBranchConfig("base", "acme/app", "main"),
			head: ptrConfig(connectedBranchConfig("head", "acme/app", "release")),
		},
		"same repo in a different url format": {
			base: connectedBranchConfig("base", "git@github.com:acme/app.git", "main"),
			head: ptrConfig(connectedBranchConfig("head", "https://github.com/Acme/App", "main")),
		},
		"same repo moved from connected to public git": {
			base: connectedBranchConfig("base", "acme/app", "main"),
			head: &app.AppBranchConfig{PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "https://github.com/acme/app.git", Branch: "main"}},
		},
		"different repo": {
			base: connectedBranchConfig("base", "acme/old", "main"),
			head: ptrConfig(connectedBranchConfig("head", "acme/new", "main")),
			want: true,
		},
		"different repo moved from public git to connected": {
			base: app.AppBranchConfig{PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "https://github.com/acme/old.git"}},
			head: ptrConfig(connectedBranchConfig("head", "acme/new", "main")),
			want: true,
		},
		"base config without a vcs config": {
			base: app.AppBranchConfig{ID: "base"},
			head: ptrConfig(connectedBranchConfig("head", "acme/new", "main")),
		},
		"nil head config": {
			base: connectedBranchConfig("base", "acme/old", "main"),
		},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tt.want, branchRepoChanged(tt.base, tt.head))
		})
	}
}

func ptrConfig(cfg app.AppBranchConfig) *app.AppBranchConfig {
	return &cfg
}
