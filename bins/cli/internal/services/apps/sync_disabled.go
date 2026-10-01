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
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	disableAppSyncFeature = "disable-app-sync"

	migrationBranchFile  = "branch.toml"
	migrationBranchesDir = "branches"

	installChoiceAll  = "all"
	installChoiceNone = "none"

	gitProbeTimeout = 3 * time.Second
)

var errMigrationCancelled = errors.New("app branch migration cancelled")

// printedErr marks an error a reused command already rendered, so the wizard
// does not print it a second time.
type printedErr struct{ error }

func (e printedErr) Unwrap() error { return e.error }

func appSyncDisabledErr() error {
	return &ui.CLIUserError{Msg: "`nuon apps sync` is disabled for this org; app config now ships through app branches. " +
		"Rerun `nuon apps sync` in an interactive terminal to migrate, or add branches/<name>.toml and run `nuon branches sync --file branches/<name>.toml`."}
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

// handleAppSyncDisabled is terminal for `nuon apps sync`: the caller must not
// go on to parse or upload the app config.
func (s *Service) handleAppSyncDisabled(ctx context.Context, dir, appID string, opts SyncOptions) error {
	if !s.appSyncWizardAvailable(opts) {
		return ui.PrintError(appSyncDisabledErr())
	}

	ok, err := bubbles.InlineConfirm("This will migrate you to app branches. Would you like to continue?", true, true)
	if err != nil || !ok {
		ui.PrintLn("migration cancelled")
		return nil
	}

	err = s.runAppBranchMigration(ctx, dir, appID)
	if errors.Is(err, errMigrationCancelled) {
		ui.PrintLn("migration cancelled")
		return nil
	}
	var printed printedErr
	if errors.As(err, &printed) {
		return err
	}
	if err != nil {
		return ui.PrintError(err)
	}
	return nil
}

func (s *Service) runAppBranchMigration(ctx context.Context, dir, appID string) error {
	if err := refuseRootBranchToml(dir); err != nil {
		return err
	}

	source, err := migrationSourceFromGit(ctx, dir)
	if err != nil {
		return err
	}
	connected, err := s.connectedRepoFullName(ctx, source.Repo)
	if err != nil {
		return err
	}
	public := connected == ""
	if !public {
		source.Repo = connected
	}

	path, err := migrationBranchConfigPath(dir, source.Branch)
	if err != nil {
		return err
	}
	rel := branchConfigDisplayPath(dir, path)

	remotes, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return fmt.Errorf("unable to list app branches: %w", err)
	}
	if existing := findBranchByName(remotes, source.Branch); existing != nil && existing.ManagedBy != appBranchManagedByConfig {
		return &ui.CLIUserError{Msg: fmt.Sprintf(
			"app branch %q already exists and is managed manually, so %s cannot take it over",
			source.Branch, rel,
		)}
	}

	cfg := newMigrationBranchConfig(source.Branch, source.Repo, source.Directory, source.Branch, public)
	rendered, err := renderBranchConfigTOML(cfg)
	if err != nil {
		return err
	}
	if public {
		ui.PrintLn(rel + " (public repo)")
	} else {
		ui.PrintLn(rel)
	}
	fmt.Println(strings.TrimRight(string(rendered), "\n"))

	_, statErr := os.Stat(path)
	exists := statErr == nil
	if statErr != nil && !os.IsNotExist(statErr) {
		return fmt.Errorf("unable to check %s: %w", path, statErr)
	}
	prompt := fmt.Sprintf("Write %s and sync it?", rel)
	defaultYes := true
	if exists {
		prompt = fmt.Sprintf("Overwrite %s and sync it?", rel)
		defaultYes = false
	}
	ok, err := bubbles.InlineConfirm(prompt, defaultYes, true)
	if err != nil || !ok {
		return errMigrationCancelled
	}
	if err := writeBranchConfigBytes(path, rendered); err != nil {
		return err
	}
	ui.PrintSuccess("wrote " + rel)

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
	branch := findBranchByName(refreshed, source.Branch)
	if branch == nil {
		return fmt.Errorf("app branch %q was not found after syncing %s", source.Branch, rel)
	}

	if err := s.migrateInstallsToBranch(ctx, appID, branch); err != nil {
		return err
	}

	ui.PrintSuccess(fmt.Sprintf("migrated to app branch %q; commit and push %s to finish", branch.Name, rel))
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

func newMigrationBranchConfig(name, repo, directory, gitBranch string, public bool) *config.AppBranchConfig {
	cfg := &config.AppBranchConfig{
		Name: name,
		Run:  &config.AppBranchRunConfig{Mode: string(models.AppAppBranchRunModePush)},
	}
	if public {
		cfg.PublicRepo = &config.PublicRepoConfig{
			Repo:      repo,
			Directory: directory,
			Branch:    gitBranch,
		}
		return cfg
	}
	cfg.ConnectedRepo = &config.ConnectedRepoConfig{
		Repo:      repo,
		Directory: directory,
		Branch:    gitBranch,
	}
	return cfg
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
	return writeBranchConfigBytes(path, byts)
}

func writeBranchConfigBytes(path string, byts []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("unable to create %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, byts, 0o644); err != nil {
		return fmt.Errorf("unable to write %s: %w", path, err)
	}
	return nil
}

func migrationBranchConfigPath(dir, name string) (string, error) {
	parts, err := branchConfigParts(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(append([]string{dir}, parts...)...), nil
}

func branchConfigParts(name string) ([]string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, &ui.CLIUserError{Msg: "git branch name is empty, so branches/<name>.toml cannot be written"}
	}
	segs := strings.Split(name, "/")
	parts := make([]string, 0, len(segs)+1)
	parts = append(parts, migrationBranchesDir)
	for i, seg := range segs {
		if seg == "" || seg == "." || seg == ".." || strings.ContainsAny(seg, `\:`) {
			return nil, &ui.CLIUserError{Msg: fmt.Sprintf("git branch %q cannot be used as a branches/ filename", name)}
		}
		if i == len(segs)-1 {
			seg += ".toml"
		}
		parts = append(parts, seg)
	}
	return parts, nil
}

func branchConfigDisplayPath(dir, path string) string {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}

func refuseRootBranchToml(dir string) error {
	path := filepath.Join(dir, migrationBranchFile)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("unable to check %s: %w", path, err)
	}
	if info.IsDir() {
		return nil
	}
	return &ui.CLIUserError{Msg: fmt.Sprintf(
		"%s already has branch.toml. Remove it before writing branches/; the app parser rejects both together. Sync the existing file with `nuon branches sync --file %s`.",
		dir, path,
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
			"could not read a GitHub remote.origin.url from %s. branch.toml needs a GitHub repo.",
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
	return matchConnectedRepo(origin, repos), nil
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

	if runs, err := s.api.GetAppBranchRuns(ctx, appID, branch.ID); err == nil && len(runs) == 0 {
		ui.PrintWarning(fmt.Sprintf("app branch %q has no runs yet; the API refuses to move installs until a branch run has completed", branch.Name))
	}

	add, err := bubbles.InlineConfirm(fmt.Sprintf("Add installs to app branch %q?", branch.Name), false, true)
	if err != nil || !add {
		ui.PrintLn("installs were not moved")
		return nil
	}

	var moved []*models.AppInstall
	var failed []installMoveFailure
	for len(pending) > 0 {
		choice, err := bubbles.SelectFromItems(
			fmt.Sprintf("Add installs to app branch %q", branch.Name),
			installSelectItems(pending),
			true,
		)
		if err != nil || choice == installChoiceNone {
			if len(moved) == 0 && len(failed) == 0 {
				ui.PrintLn("installs were not moved")
			}
			break
		}
		batch, rest := installsForChoice(pending, choice)
		if len(batch) == 0 {
			break
		}
		pending = rest
		result := moveInstallsToBranch(ctx, batch, branch.ID, func(ctx context.Context, installID, branchID string) error {
			_, err := s.api.MoveInstallToAppBranch(ctx, installID, branchID, "")
			return err
		})
		for _, inst := range result.Moved {
			ui.PrintSuccess("moved " + installLabel(inst))
		}
		moved = append(moved, result.Moved...)
		failed = append(failed, result.Failed...)
	}
	if len(failed) == 0 {
		return nil
	}

	msgs := make([]string, 0, len(failed))
	for _, f := range failed {
		msgs = append(msgs, fmt.Sprintf("%s: %s", installLabel(f.Install), apiErrorMessage(f.Err)))
	}
	return &ui.CLIUserError{Msg: fmt.Sprintf(
		"moved %d of %d install(s) to app branch %q; failed:\n  %s",
		len(moved), len(moved)+len(failed), branch.Name, strings.Join(msgs, "\n  "),
	)}
}

func installSelectItems(pending []*models.AppInstall) []bubbles.SelectorItem {
	items := []bubbles.SelectorItem{
		bubbles.NewSelectorItem("All", "", installChoiceAll),
		bubbles.NewSelectorItem("None", "", installChoiceNone),
	}
	for _, inst := range pending {
		items = append(items, bubbles.NewSelectorItem(installLabel(inst), "", inst.ID))
	}
	return items
}

func installsForChoice(pending []*models.AppInstall, choice string) (batch, rest []*models.AppInstall) {
	if choice == "" || choice == installChoiceNone {
		return nil, pending
	}
	if choice == installChoiceAll {
		return pending, nil
	}
	rest = make([]*models.AppInstall, 0, len(pending))
	for _, inst := range pending {
		if inst.ID == choice && len(batch) == 0 {
			batch = []*models.AppInstall{inst}
			continue
		}
		rest = append(rest, inst)
	}
	if len(batch) == 0 {
		return nil, pending
	}
	return batch, rest
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
