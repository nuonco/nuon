package activities

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestValidateChangedFilePathsInput(t *testing.T) {
	sha := strings.Repeat("a", 40)
	zero := 0
	tests := []struct {
		name    string
		input   FetchChangedFilePathsInput
		wantErr bool
	}{
		{name: "commit only", input: FetchChangedFilePathsInput{CommitSHA: sha}},
		{name: "sha256 commit", input: FetchChangedFilePathsInput{CommitSHA: strings.Repeat("b", 64)}},
		{name: "compare", input: FetchChangedFilePathsInput{CommitSHA: sha, BaseSHA: strings.Repeat("c", 40)}},
		{name: "base branch", input: FetchChangedFilePathsInput{CommitSHA: sha, BaseBranch: "release/v1"}},
		{name: "short sha", input: FetchChangedFilePathsInput{CommitSHA: "abc123"}, wantErr: true},
		{name: "uppercase sha", input: FetchChangedFilePathsInput{CommitSHA: strings.Repeat("A", 40)}, wantErr: true},
		{name: "path in commit", input: FetchChangedFilePathsInput{CommitSHA: "../../orgs/acme"}, wantErr: true},
		{name: "path in base sha", input: FetchChangedFilePathsInput{CommitSHA: sha, BaseSHA: "main/../x"}, wantErr: true},
		{name: "traversal base branch", input: FetchChangedFilePathsInput{CommitSHA: sha, BaseBranch: "../../orgs"}, wantErr: true},
		{name: "query in base branch", input: FetchChangedFilePathsInput{CommitSHA: sha, BaseBranch: "main?x=1"}, wantErr: true},
		{name: "encoded base branch", input: FetchChangedFilePathsInput{CommitSHA: sha, BaseBranch: "main%2F..%2Fx"}, wantErr: true},
		{name: "zero pr number", input: FetchChangedFilePathsInput{CommitSHA: sha, PRNumber: &zero}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateChangedFilePathsInput(&tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestIsSafeGitRef(t *testing.T) {
	for _, ref := range []string{"v1.2.3", "release/v1", "acme/v0.0.1"} {
		require.True(t, isSafeGitRef(ref), ref)
	}
	for _, ref := range []string{"", "/v1", "v1/", "../v1", "v1..v2", `a\b`, "v1?x", "v1#x", "v1%2e", "v1 x", "v1\n"} {
		require.False(t, isSafeGitRef(ref), ref)
	}
}

func TestPushMatchesRunPathsFailsOpenWithoutComparableSHAs(t *testing.T) {
	a := &Activities{}
	config := &app.AppBranchConfig{
		ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{Directory: "apps/acme"},
	}
	runConfig := app.AppBranchRunConfig{Mode: app.AppBranchRunModePush, Paths: []string{"apps/acme/**"}}
	sha := strings.Repeat("a", 40)

	for name, req := range map[string]*TriggerAppBranchRunFromVCSPushRequest{
		"new branch":      {HeadSHA: sha, BaseSHA: strings.Repeat("0", 40), ChangedFiles: []string{"README.md"}},
		"missing base":    {HeadSHA: sha, ChangedFiles: []string{"README.md"}},
		"invalid head":    {HeadSHA: "../x", BaseSHA: sha, ChangedFiles: []string{"README.md"}},
		"invalid base":    {HeadSHA: sha, BaseSHA: "main", ChangedFiles: []string{"README.md"}},
		"no vcs config":   {HeadSHA: sha, BaseSHA: strings.Repeat("b", 40)},
		"empty push data": {},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := config
			if name == "no vcs config" {
				cfg = &app.AppBranchConfig{}
			}
			require.True(t, a.pushMatchesRunPaths(context.Background(), cfg, runConfig, req))
		})
	}

	t.Run("no filter at repo root", func(t *testing.T) {
		root := &app.AppBranchConfig{ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{Directory: "."}}
		require.True(t, a.pushMatchesRunPaths(context.Background(), root, app.AppBranchRunConfig{Mode: app.AppBranchRunModePush}, &TriggerAppBranchRunFromVCSPushRequest{}))
	})
}
