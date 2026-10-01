package apps

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	cliconfig "github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestMigrationBranchConfigPath(t *testing.T) {
	require.Equal(t, filepath.Join("apps", "acme", "branch.toml"), migrationBranchConfigPath(filepath.Join("apps", "acme")))
}

func TestRelativeRepoDirectory(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "apps", "platform")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	got, err := relativeRepoDirectory(root, nested)
	require.NoError(t, err)
	require.Equal(t, "apps/platform", got)

	got, err = relativeRepoDirectory(root, root)
	require.NoError(t, err)
	require.Equal(t, ".", got)

	_, err = relativeRepoDirectory(nested, root)
	require.Error(t, err)
}

func TestMatchConnectedRepo(t *testing.T) {
	repos := []*models.ServiceVCSConnectionRepo{
		{FullName: "acme/other"},
		nil,
		{FullName: "Acme/Platform"},
	}
	require.Equal(t, "Acme/Platform", matchConnectedRepo("acme/platform", repos))
	require.Empty(t, matchConnectedRepo("acme/missing", repos))
}

func TestMigrationBranchConfigRoundTrips(t *testing.T) {
	cfg := newMigrationBranchConfig("default", "acme/platform", ".", "main")

	byts, err := renderBranchConfigTOML(cfg)
	require.NoError(t, err)

	parsed, err := parse.ParseAppBranchConfig(bytes.NewReader(byts))
	require.NoError(t, err)
	require.Equal(t, "default", parsed.Name)
	require.Equal(t, &config.ConnectedRepoConfig{Repo: "acme/platform", Directory: ".", Branch: "main"}, parsed.ConnectedRepo)
	require.NotNil(t, parsed.Run)
	require.Equal(t, "push", parsed.Run.Mode)
	require.Nil(t, parsed.PublicRepo)
	require.Empty(t, parsed.InstallGroups)
	require.Nil(t, parsed.Preview)
}

func TestWriteBranchConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := migrationBranchConfigPath(dir)

	require.NoError(t, writeBranchConfigFile(path, newMigrationBranchConfig("main", "acme/platform", "apps/platform", "main")))

	cfg, err := parse.ParseAppBranchConfigFile(path)
	require.NoError(t, err)
	require.Equal(t, "main", cfg.Name)
	require.Equal(t, "apps/platform", cfg.ConnectedRepo.Directory)
}

func TestCheckEmbeddedBranches(t *testing.T) {
	tests := map[string]struct {
		cfg     *config.AppConfig
		wantErr bool
	}{
		"nil config":  {cfg: nil},
		"no branches": {cfg: &config.AppConfig{}},
		"branch.toml": {cfg: &config.AppConfig{Branch: &config.AppBranchConfig{Name: "default"}}, wantErr: true},
		"branches/":   {cfg: &config.AppConfig{Branches: []*config.AppBranchConfig{{Name: "a"}}}, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := checkEmbeddedBranches(tc.cfg, "acme")
			if !tc.wantErr {
				require.NoError(t, err)
				return
			}
			var userErr *ui.CLIUserError
			require.ErrorAs(t, err, &userErr)
			require.Contains(t, userErr.Msg, "nuon branches sync")
		})
	}
}

// Embedded branches in an app config dir are parsed by the same loader the
// wizard's branches/<name>.toml output lands in, so the check must see them.
func TestCheckEmbeddedBranchesFromParsedDir(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "metadata.toml"), []byte("version = \"v1\"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sandbox.toml"), []byte(`terraform_version = "1.11.3"
[public_repo]
repo = "acme/aws-eks-sandbox"
directory = "."
branch = "main"
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "runner.toml"), []byte(`runner_type = "aws"
helm_driver = "configmap"
init_script_url = "https://example.com/init.sh"
`), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(dir, "components"), 0o755))
	require.NoError(t, writeBranchConfigFile(migrationBranchConfigPath(dir), newMigrationBranchConfig("default", "acme/platform", ".", "main")))

	cfg, err := parse.ParseDir(context.Background(), parse.ParseConfig{
		Dirname:       dir,
		FileProcessor: func(_ string, obj map[string]any) map[string]any { return obj },
	})
	require.NoError(t, err)
	require.Error(t, checkEmbeddedBranches(cfg, dir))
}

type appSyncDisabledAPI struct {
	nuon.Client
	cliConfig *models.ServiceCLIConfig
}

func (a *appSyncDisabledAPI) GetAppBranches(_ context.Context, _ string, _ *models.GetPaginatedQuery) ([]*models.AppAppBranch, bool, error) {
	return nil, false, nil
}

func (a *appSyncDisabledAPI) GetCLIConfig(_ context.Context) (*models.ServiceCLIConfig, error) {
	return a.cliConfig, nil
}

func TestHandleAppSyncDisabledNonInteractive(t *testing.T) {
	tests := map[string]struct {
		interactive bool
		opts        SyncOptions
	}{
		"no tty":      {interactive: false},
		"json on tty": {interactive: true, opts: SyncOptions{PrintJSON: true}},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			api := &appSyncDisabledAPI{
				cliConfig: &models.ServiceCLIConfig{DashboardURL: "https://app.nuon.co"},
			}
			s := &Service{
				api: api,
				cfg: &cliconfig.Config{Interactive: tc.interactive, OrgID: "org-1"},
			}

			err := s.handleAppSyncDisabled(context.Background(), dir, "app-1", tc.opts)

			var userErr *ui.CLIUserError
			require.ErrorAs(t, err, &userErr)
			require.Contains(t, userErr.Msg, "disabled")
			require.Contains(t, userErr.Msg, "nuon branches sync")
			require.Contains(t, userErr.Msg, "move installs onto that branch")
			require.Contains(t, userErr.Msg, "No app branches exist yet.")
			entries, readErr := os.ReadDir(dir)
			require.NoError(t, readErr)
			require.Empty(t, entries)
		})
	}
}

func TestFindBranchByName(t *testing.T) {
	remotes := []*models.AppAppBranch{
		{ID: "b1", Name: "default"},
		nil,
		{ID: "b3", Name: "manual", ManagedBy: appBranchManagedByManually},
	}
	require.Equal(t, "b3", findBranchByName(remotes, "manual").ID)
	require.Nil(t, findBranchByName(remotes, "missing"))
}

func TestRenderAppSyncBranchGuides(t *testing.T) {
	got := renderAppSyncBranchGuides([]appSyncBranchGuide{{
		Name:         "production",
		ID:           "br-1",
		ManagedBy:    appBranchManagedByConfig,
		DashboardURL: "https://app.nuon.co/org-1/apps/app-1/branches/br-1",
		ConfigPath:   "branch.toml",
		Config:       "name = 'production'\n[run]\nmode = 'push'",
	}})

	require.Contains(t, got, "production (br-1, managed by config)")
	require.Contains(t, got, "https://app.nuon.co/org-1/apps/app-1/branches/br-1")
	require.Contains(t, got, "nuon branches sync --file branch.toml")
	require.Contains(t, got, "name = 'production'")
}

func TestGithubRepoFromRemoteURL(t *testing.T) {
	tests := map[string]string{
		"git@github.com:acme/platform.git":         "acme/platform",
		"https://github.com/acme/platform.git":     "acme/platform",
		"https://github.com/acme/platform":         "acme/platform",
		"https://user@github.com/acme/platform/":   "acme/platform",
		"ssh://git@github.com/acme/platform.git":   "acme/platform",
		"https://gitlab.example.com/acme/platform": "",
		"": "",
	}
	for in, want := range tests {
		t.Run(in, func(t *testing.T) {
			require.Equal(t, want, githubRepoFromRemoteURL(in))
		})
	}
}

func TestInstallsToMove(t *testing.T) {
	installs := []*models.AppInstall{
		{ID: "i1", AppBranchID: "br-1"},
		{ID: "i2", AppBranchID: "br-other"},
		{ID: "i3"},
		nil,
	}
	got := installsToMove(installs, "br-1")
	require.Len(t, got, 2)
	require.Equal(t, "i2", got[0].ID)
	require.Equal(t, "i3", got[1].ID)
}

func TestMoveInstallsToBranchReportsPartialFailures(t *testing.T) {
	installs := []*models.AppInstall{{ID: "i1"}, {ID: "i2"}, {ID: "i3"}}
	var calls []string

	result := moveInstallsToBranch(context.Background(), installs, "br-1", func(_ context.Context, installID, branchID string) error {
		require.Equal(t, "br-1", branchID)
		calls = append(calls, installID)
		if installID == "i2" {
			return errors.New("no deployable run")
		}
		return nil
	})

	require.Equal(t, []string{"i1", "i2", "i3"}, calls)
	require.Len(t, result.Moved, 2)
	require.Len(t, result.Failed, 1)
	require.Equal(t, "i2", result.Failed[0].Install.ID)
	require.EqualError(t, result.Failed[0].Err, "no deployable run")
}
