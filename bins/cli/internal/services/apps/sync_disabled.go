package apps

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/bins/cli/internal/paginate"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	disableAppSyncFeature = "disable-app-sync"

	migrationBranchFile  = "branch.toml"
	migrationBranchesDir = "branches"

	fileChoiceUseExisting = "use-existing"
	fileChoiceOverwrite   = "overwrite"
	fileChoiceCancel      = "cancel"

	gitProbeTimeout = 3 * time.Second
)

var errMigrationCancelled = errors.New("app branch migration cancelled")

// printedErr marks an error a reused command already rendered, so the wizard
// does not print it a second time.
type printedErr struct{ error }

func (e printedErr) Unwrap() error { return e.error }

func appSyncDisabledErr(guidance string) error {
	msg := "`nuon apps sync` is disabled for this org: app config now ships through config-managed app branches. " +
		"Add a branch.toml next to the app config (name, [connected_repo], and [run] mode = \"push\"), " +
		"apply it with `nuon branches sync --file branch.toml`, then move installs onto that branch. " +
		"Rerun `nuon apps sync` in an interactive terminal to be walked through the migration."
	if guidance != "" {
		msg += "\n\n" + guidance
	}
	return &ui.CLIUserError{Msg: msg}
}

func embeddedBranchesErr(dir string) error {
	return &ui.CLIUserError{Msg: fmt.Sprintf(
		"the app config in %s declares app branches (branch.toml or branches/). Syncing app branches through `nuon apps sync` is deprecated: "+
			"sync them with `nuon branches sync --file %s`, and remove them from the app config directory to keep using `nuon apps sync`.",
		dir, filepath.Join(dir, migrationBranchesDir),
	)}
}

func checkEmbeddedBranches(cfg *config.AppConfig, dir string) error {
	if cfg == nil {
		return nil
	}
	if cfg.Branch != nil || len(cfg.Branches) > 0 {
		return embeddedBranchesErr(dir)
	}
	return nil
}

func (s *Service) appSyncWizardAvailable(opts SyncOptions) bool {
	return s.cfg.Interactive && !opts.PrintJSON && !agentmode.Enabled()
}

// handleAppSyncDisabled is terminal for `nuon apps sync`: it either returns the
// deprecation error or runs the migration wizard, and the caller must not go on
// to parse or upload the app config either way.
func (s *Service) handleAppSyncDisabled(ctx context.Context, dir, appID string, opts SyncOptions) error {
	guides, guideErr := s.appSyncBranchGuides(ctx, dir, appID)
	guidance := renderAppSyncBranchGuides(guides)
	if guideErr != nil {
		guidance = "Unable to load existing app branches: " + guideErr.Error()
	}

	if !s.appSyncWizardAvailable(opts) {
		return ui.PrintError(appSyncDisabledErr(guidance))
	}

	printMigrationIntro()
	if guidance != "" {
		ui.PrintLn(guidance)
	}
	start, err := bubbles.InlineConfirm("Set up an app branch now?", true, true)
	if err != nil || !start {
		return ui.PrintError(appSyncDisabledErr(guidance))
	}

	err = s.runAppBranchMigration(ctx, dir, appID)
	if errors.Is(err, errMigrationCancelled) {
		ui.PrintLn("migration cancelled; nothing further was changed")
		return ui.PrintError(appSyncDisabledErr(guidance))
	}
	var printed printedErr
	if errors.As(err, &printed) {
		return err
	}
	if err != nil {
		return ui.PrintError(err)
	}

	updatedGuides, updatedErr := s.appSyncBranchGuides(ctx, dir, appID)
	if updatedErr == nil {
		guidance = renderAppSyncBranchGuides(updatedGuides)
	}
	return ui.PrintError(appSyncDisabledErr(guidance))
}

type appSyncBranchGuide struct {
	Name         string
	ID           string
	ManagedBy    string
	DashboardURL string
	ConfigPath   string
	Config       string
	ConfigError  string
}

func (s *Service) appSyncBranchGuides(ctx context.Context, dir, appID string) ([]appSyncBranchGuide, error) {
	branches, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to list app branches: %w", err)
	}

	dashboardURL := ""
	if cliCfg, cfgErr := s.api.GetCLIConfig(ctx); cfgErr == nil {
		dashboardURL = strings.TrimRight(cliCfg.DashboardURL, "/")
	}

	localPath, localName := localBranchConfig(dir)

	resolver := newBranchNameResolver(s.api, appID)
	guides := make([]appSyncBranchGuide, 0, len(branches))
	for _, branch := range branches {
		if branch == nil {
			continue
		}
		guide := appSyncBranchGuide{
			Name:      branch.Name,
			ID:        branch.ID,
			ManagedBy: branch.ManagedBy,
		}
		if dashboardURL != "" {
			guide.DashboardURL = fmt.Sprintf("%s/%s/apps/%s/branches/%s", dashboardURL, s.cfg.OrgID, appID, branch.ID)
		}

		if localPath != "" && branch.Name == localName {
			guide.ConfigPath = localPath
		}

		latest, latestErr := s.latestBranchConfig(ctx, appID, branch.ID)
		if latestErr != nil {
			guide.ConfigError = latestErr.Error()
			guides = append(guides, guide)
			continue
		}
		normalized, normalizeErr := normalizeRemoteBranch(ctx, resolver, branch.Name, latest)
		if normalizeErr != nil {
			guide.ConfigError = normalizeErr.Error()
			guides = append(guides, guide)
			continue
		}
		byts, renderErr := renderBranchConfigTOML(normalized)
		if renderErr != nil {
			guide.ConfigError = renderErr.Error()
		} else {
			guide.Config = strings.TrimSpace(string(byts))
		}
		guides = append(guides, guide)
	}
	return guides, nil
}

func renderAppSyncBranchGuides(guides []appSyncBranchGuide) string {
	if len(guides) == 0 {
		return "No app branches exist yet."
	}

	var out strings.Builder
	out.WriteString("Existing app branches:\n")
	for _, guide := range guides {
		fmt.Fprintf(&out, "\n- %s (%s, managed by %s)\n", guide.Name, guide.ID, guide.ManagedBy)
		if guide.DashboardURL != "" {
			fmt.Fprintf(&out, "  Dashboard: %s\n", guide.DashboardURL)
		}
		if guide.ConfigPath != "" {
			fmt.Fprintf(&out, "  Local config: %s\n", guide.ConfigPath)
			fmt.Fprintf(&out, "  Sync: nuon branches sync --file %s\n", guide.ConfigPath)
		}
		if guide.ConfigError != "" {
			fmt.Fprintf(&out, "  Config unavailable: %s\n", guide.ConfigError)
			continue
		}
		if guide.Config != "" {
			out.WriteString("  Shareable config:\n")
			for _, line := range strings.Split(guide.Config, "\n") {
				fmt.Fprintf(&out, "    %s\n", line)
			}
		}
	}
	return strings.TrimRight(out.String(), "\n")
}

func printMigrationIntro() {
	lines := []string{
		"`nuon apps sync` is disabled for this org.",
		"App config now ships through a config-managed app branch: a branch file in your repo points",
		"Nuon at a repo, directory and git branch, and pushes to that git branch run the app branch.",
		"",
		"This wizard will:",
		"  1. write branch.toml in this app config directory from the git checkout and the connected GitHub repo",
		"  2. create or update that app branch with `nuon branches sync`",
		"  3. ask whether to move this app's installs onto the branch (no by default)",
		"",
		"It will not sync this app config; commit and push branch.toml afterwards.",
	}
	for _, l := range lines {
		fmt.Println(styles.TextDim.Render("  " + l))
	}
}

func (s *Service) runAppBranchMigration(ctx context.Context, dir, appID string) error {
	if err := refuseBranchesDirectory(dir); err != nil {
		return err
	}

	source, err := migrationSourceFromGit(ctx, dir)
	if err != nil {
		return err
	}
	repo, err := s.connectedRepoFullName(ctx, source.Repo)
	if err != nil {
		return err
	}
	source.Repo = repo

	remotes, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return fmt.Errorf("unable to list app branches: %w", err)
	}
	if existing := findBranchByName(remotes, source.Branch); existing != nil && existing.ManagedBy != appBranchManagedByConfig {
		return &ui.CLIUserError{Msg: fmt.Sprintf(
			"app branch %q already exists and is managed manually, so branch.toml cannot take it over",
			source.Branch,
		)}
	}

	ui.PrintLn(fmt.Sprintf("branch.toml will track %s (%s) on git branch %s", source.Repo, source.Directory, source.Branch))

	path := migrationBranchConfigPath(dir)
	write := true
	branchName := source.Branch
	if _, statErr := os.Stat(path); statErr == nil {
		choice, err := chooseExistingFileAction(path)
		if err != nil {
			return err
		}
		switch choice {
		case fileChoiceCancel:
			return errMigrationCancelled
		case fileChoiceUseExisting:
			write = false
		}
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("unable to check %s: %w", path, statErr)
	}

	if write {
		cfg := newMigrationBranchConfig(source.Branch, source.Repo, source.Directory, source.Branch)
		if err := writeBranchConfigFile(path, cfg); err != nil {
			return err
		}
		ui.PrintSuccess("wrote " + path)
	} else {
		existingCfg, err := parse.ParseAppBranchConfigFile(path)
		if err != nil {
			return err
		}
		branchName = existingCfg.Name
	}

	if err := s.SyncBranches(ctx, SyncBranchesOptions{
		Path:    path,
		AppID:   appID,
		Confirm: true,
	}); err != nil {
		return printedErr{err}
	}

	refreshed, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return fmt.Errorf("unable to list app branches: %w", err)
	}
	branch := findBranchByName(refreshed, branchName)
	if branch == nil {
		return fmt.Errorf("app branch %q was not found after syncing %s", branchName, path)
	}

	if err := s.migrateInstallsToBranch(ctx, appID, branch); err != nil {
		return err
	}

	ui.PrintLn("next: commit branch.toml and push it; pushes to the tracked git branch now run this app branch.")
	ui.PrintLn("`nuon apps sync` stays disabled for this org; use `nuon branches sync --file branch.toml` to change branch settings.")
	return nil
}

func findBranchByName(remotes []*models.AppAppBranch, name string) *models.AppAppBranch {
	for _, b := range remotes {
		if b != nil && b.Name == name {
			return b
		}
	}
	return nil
}

func chooseExistingFileAction(path string) (string, error) {
	items := []bubbles.SelectorItem{
		bubbles.NewSelectorItem("Use the existing file as is", "", fileChoiceUseExisting),
		bubbles.NewSelectorItem("Overwrite it", "", fileChoiceOverwrite),
		bubbles.NewSelectorItem("Cancel", "", fileChoiceCancel),
	}
	choice, err := bubbles.SelectFromItems(fmt.Sprintf("%s already exists", path), items, true)
	if err != nil {
		return "", errMigrationCancelled
	}
	if choice == fileChoiceOverwrite {
		ok, err := bubbles.InlineConfirm(fmt.Sprintf("Overwrite %s?", path), false, true)
		if err != nil || !ok {
			return "", errMigrationCancelled
		}
	}
	return choice, nil
}

func newMigrationBranchConfig(name, repo, directory, gitBranch string) *config.AppBranchConfig {
	return &config.AppBranchConfig{
		Name: name,
		ConnectedRepo: &config.ConnectedRepoConfig{
			Repo:      repo,
			Directory: directory,
			Branch:    gitBranch,
		},
		Run: &config.AppBranchRunConfig{Mode: string(models.AppAppBranchRunModePush)},
	}
}

func renderBranchConfigTOML(cfg *config.AppBranchConfig) ([]byte, error) {
	byts, err := toml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to render branch config: %w", err)
	}
	return byts, nil
}

func writeBranchConfigFile(path string, cfg *config.AppBranchConfig) error {
	byts, err := renderBranchConfigTOML(cfg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("unable to create %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, byts, 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", path, err)
	}
	return nil
}

func migrationBranchConfigPath(dir string) string {
	return filepath.Join(dir, migrationBranchFile)
}

func localBranchConfig(dir string) (path, name string) {
	path = migrationBranchConfigPath(dir)
	if _, err := os.Stat(path); err != nil {
		return "", ""
	}
	cfg, err := parse.ParseAppBranchConfigFile(path)
	if err != nil || cfg == nil {
		return "", ""
	}
	return path, cfg.Name
}

func refuseBranchesDirectory(dir string) error {
	branchesDir := filepath.Join(dir, migrationBranchesDir)
	info, err := os.Stat(branchesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to check %s: %w", branchesDir, err)
	}
	if !info.IsDir() {
		return nil
	}
	return &ui.CLIUserError{Msg: fmt.Sprintf(
		"%s already has a branches/ directory. Sync those files with `nuon branches sync --file %s` instead of writing branch.toml alongside them.",
		dir, branchesDir,
	)}
}

type migrationGitSource struct {
	Repo      string
	Directory string
	Branch    string
}

func migrationSourceFromGit(ctx context.Context, dir string) (*migrationGitSource, error) {
	repo := gitOriginRepo(ctx, dir)
	if repo == "" {
		return nil, &ui.CLIUserError{Msg: fmt.Sprintf(
			"could not read a GitHub remote.origin.url from %s. branch.toml needs a connected GitHub repo.",
			dir,
		)}
	}
	branch := gitCurrentBranch(ctx, dir)
	if branch == "" || branch == "HEAD" {
		return nil, &ui.CLIUserError{Msg: "check out a git branch before migrating; a detached HEAD has no branch name to write into branch.toml"}
	}
	directory, err := gitConfigDirectory(ctx, dir)
	if err != nil {
		return nil, err
	}
	return &migrationGitSource{Repo: repo, Directory: directory, Branch: branch}, nil
}

func gitConfigDirectory(ctx context.Context, dir string) (string, error) {
	toplevel := gitOutput(ctx, dir, "rev-parse", "--show-toplevel")
	if toplevel == "" {
		return "", &ui.CLIUserError{Msg: fmt.Sprintf("%s is not a git checkout, so branch.toml cannot be generated from it", dir)}
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s: %w", dir, err)
	}
	return relativeRepoDirectory(toplevel, absDir)
}

func relativeRepoDirectory(toplevel, dir string) (string, error) {
	absTop, err := resolvePath(toplevel)
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s: %w", toplevel, err)
	}
	absDir, err := resolvePath(dir)
	if err != nil {
		return "", fmt.Errorf("unable to resolve %s: %w", dir, err)
	}
	rel, err := filepath.Rel(absTop, absDir)
	if err != nil {
		return "", fmt.Errorf("unable to locate %s within %s: %w", dir, toplevel, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", &ui.CLIUserError{Msg: fmt.Sprintf("%s is not inside the git checkout at %s", dir, toplevel)}
	}
	if rel == "." {
		return ".", nil
	}
	return filepath.ToSlash(rel), nil
}

func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs, nil
	}
	return resolved, nil
}

func (s *Service) connectedRepoFullName(ctx context.Context, origin string) (string, error) {
	connections, err := paginate.All(func(offset, limit int) ([]*models.AppVCSConnection, bool, error) {
		return s.api.GetVCSConnections(ctx, &models.GetPaginatedQuery{Offset: offset, Limit: limit})
	})
	if err != nil {
		return "", fmt.Errorf("unable to list vcs connections: %w", err)
	}

	var repos []*models.ServiceVCSConnectionRepo
	for _, conn := range connections {
		if conn == nil {
			continue
		}
		resp, err := s.api.GetVCSConnectionRepos(ctx, conn.ID)
		if err != nil {
			return "", fmt.Errorf("unable to list repos for vcs connection %s: %w", conn.ID, err)
		}
		if resp == nil {
			continue
		}
		repos = append(repos, resp.Repositories...)
	}
	if fullName := matchConnectedRepo(origin, repos); fullName != "" {
		return fullName, nil
	}
	return "", &ui.CLIUserError{Msg: fmt.Sprintf(
		"git repo %q is not on a connected GitHub installation in this org. Connect it, then rerun `nuon apps sync`.",
		origin,
	)}
}

func matchConnectedRepo(origin string, repos []*models.ServiceVCSConnectionRepo) string {
	for _, repo := range repos {
		if repo != nil && strings.EqualFold(repo.FullName, origin) {
			return repo.FullName
		}
	}
	return ""
}

var githubRemotePattern = regexp.MustCompile(`^(?:https?://|ssh://)?(?:[^@/]+@)?github\.com[:/]([^/\s]+)/([^/\s]+?)(?:\.git)?/?$`)

func githubRepoFromRemoteURL(remote string) string {
	m := githubRemotePattern.FindStringSubmatch(strings.TrimSpace(remote))
	if m == nil {
		return ""
	}
	return m[1] + "/" + m[2]
}

func gitOutput(ctx context.Context, dir string, args ...string) string {
	ctx, cancel := context.WithTimeout(ctx, gitProbeTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func gitOriginRepo(ctx context.Context, dir string) string {
	return githubRepoFromRemoteURL(gitOutput(ctx, dir, "config", "--get", "remote.origin.url"))
}

func gitCurrentBranch(ctx context.Context, dir string) string {
	return gitOutput(ctx, dir, "rev-parse", "--abbrev-ref", "HEAD")
}

func (s *Service) migrateInstallsToBranch(ctx context.Context, appID string, branch *models.AppAppBranch) error {
	installs, err := paginate.All(func(offset, limit int) ([]*models.AppInstall, bool, error) {
		return s.api.GetAppInstalls(ctx, appID, &models.GetPaginatedQuery{Offset: offset, Limit: limit})
	})
	if err != nil {
		return fmt.Errorf("unable to list app installs: %w", err)
	}

	pending := installsToMove(installs, branch.ID)
	if len(pending) == 0 {
		ui.PrintSuccess(fmt.Sprintf("all installs are already on app branch %q", branch.Name))
		return nil
	}

	fmt.Println(styles.TextDim.Render(fmt.Sprintf("  %d install(s) are not on app branch %q:", len(pending), branch.Name)))
	for _, inst := range pending {
		fmt.Println(styles.TextDim.Render("    - " + installLabel(inst)))
	}
	if runs, err := s.api.GetAppBranchRuns(ctx, appID, branch.ID); err == nil && len(runs) == 0 {
		ui.PrintWarning(fmt.Sprintf("app branch %q has no runs yet; the API refuses to move installs until a branch run has completed", branch.Name))
	}

	move, err := bubbles.InlineConfirm(fmt.Sprintf("Move %d install(s) to app branch %q?", len(pending), branch.Name), false, true)
	if err != nil || !move {
		ui.PrintLn("installs were not moved")
		return nil
	}

	result := moveInstallsToBranch(ctx, pending, branch.ID, func(ctx context.Context, installID, branchID string) error {
		_, err := s.api.MoveInstallToAppBranch(ctx, installID, branchID, "")
		return err
	})
	for _, inst := range result.Moved {
		ui.PrintSuccess("moved " + installLabel(inst))
	}
	if len(result.Failed) == 0 {
		return nil
	}

	msgs := make([]string, 0, len(result.Failed))
	for _, f := range result.Failed {
		msgs = append(msgs, fmt.Sprintf("%s: %s", installLabel(f.Install), apiErrorMessage(f.Err)))
	}
	return &ui.CLIUserError{Msg: fmt.Sprintf(
		"moved %d of %d install(s) to app branch %q; failed:\n  %s",
		len(result.Moved), len(pending), branch.Name, strings.Join(msgs, "\n  "),
	)}
}

func installsToMove(installs []*models.AppInstall, branchID string) []*models.AppInstall {
	out := make([]*models.AppInstall, 0, len(installs))
	for _, inst := range installs {
		if inst == nil || inst.AppBranchID == branchID {
			continue
		}
		out = append(out, inst)
	}
	return out
}

type moveInstallFunc func(ctx context.Context, installID, branchID string) error

type installMoveFailure struct {
	Install *models.AppInstall
	Err     error
}

type installMoveResult struct {
	Moved  []*models.AppInstall
	Failed []installMoveFailure
}

func moveInstallsToBranch(ctx context.Context, installs []*models.AppInstall, branchID string, move moveInstallFunc) installMoveResult {
	var result installMoveResult
	for _, inst := range installs {
		if err := move(ctx, inst.ID, branchID); err != nil {
			result.Failed = append(result.Failed, installMoveFailure{Install: inst, Err: err})
			continue
		}
		result.Moved = append(result.Moved, inst)
	}
	return result
}

func installLabel(inst *models.AppInstall) string {
	if inst.Name == "" {
		return inst.ID
	}
	return fmt.Sprintf("%s (%s)", inst.Name, inst.ID)
}

func apiErrorMessage(err error) string {
	if ue, ok := nuon.ToUserError(err); ok && ue.Description != "" {
		return ue.Description
	}
	if msg, ok := nuon.ToAPIError(err); ok && msg != "" {
		return msg
	}
	return err.Error()
}
