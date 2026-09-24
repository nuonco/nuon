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

	migrationDefaultBranchName = "default"
	migrationDefaultDirectory  = "."
	migrationBranchesDir       = "branches"

	migrationEnterOtherValue = "__enter_other__"
	migrationCreateNewBranch = "__create_new__"

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
		"Add a branch file (for example branches/default.toml with name, [connected_repo] and [run] mode = \"push\"), " +
		"apply it with `nuon branches sync --file branches/default.toml`, then move installs onto that branch. " +
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

		path := branchConfigFilePath(dir, branch.Name)
		if _, statErr := os.Stat(path); statErr == nil {
			guide.ConfigPath = path
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
		"  1. write branches/<name>.toml in this app config directory (asking before overwriting)",
		"  2. create or update that app branch with `nuon branches sync`",
		"  3. optionally move this app's installs onto the branch",
		"",
		"It will not sync this app config; commit and push the branch file afterwards.",
	}
	for _, l := range lines {
		fmt.Println(styles.TextDim.Render("  " + l))
	}
}

func (s *Service) runAppBranchMigration(ctx context.Context, dir, appID string) error {
	remotes, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return fmt.Errorf("unable to list app branches: %w", err)
	}

	branch, err := s.pickExistingConfigBranch(remotes)
	if err != nil {
		return err
	}

	if branch == nil {
		branch, err = s.createMigrationBranch(ctx, dir, appID, remotes)
		if err != nil {
			return err
		}
	}

	if err := s.migrateInstallsToBranch(ctx, appID, branch); err != nil {
		return err
	}

	ui.PrintLn("next: commit the branch file and push it; pushes to the tracked git branch now run this app branch.")
	ui.PrintLn("`nuon apps sync` stays disabled for this org; use `nuon branches sync --file <file>` to change branch settings.")
	return nil
}

func configManagedBranches(remotes []*models.AppAppBranch) []*models.AppAppBranch {
	out := make([]*models.AppAppBranch, 0, len(remotes))
	for _, b := range remotes {
		if b != nil && b.ManagedBy == appBranchManagedByConfig {
			out = append(out, b)
		}
	}
	return out
}

func findBranchByName(remotes []*models.AppAppBranch, name string) *models.AppAppBranch {
	for _, b := range remotes {
		if b != nil && b.Name == name {
			return b
		}
	}
	return nil
}

func (s *Service) pickExistingConfigBranch(remotes []*models.AppAppBranch) (*models.AppAppBranch, error) {
	managed := configManagedBranches(remotes)
	if len(managed) == 0 {
		return nil, nil
	}

	items := make([]bubbles.SelectorItem, 0, len(managed)+1)
	items = append(items, bubbles.NewSelectorItem("Create or update a branch from a new branch file", "", migrationCreateNewBranch))
	for _, b := range managed {
		items = append(items, bubbles.NewSelectorItem(
			"Reuse "+b.Name,
			styles.TextDim.Render("config-managed, ID: "+b.ID+" (left as is)"),
			b.ID,
		))
	}
	choice, err := bubbles.SelectFromItems("This app already has config-managed branches", items, true)
	if err != nil {
		return nil, errMigrationCancelled
	}
	if choice == migrationCreateNewBranch {
		return nil, nil
	}
	for _, b := range managed {
		if b.ID == choice {
			return b, nil
		}
	}
	return nil, nil
}

func (s *Service) createMigrationBranch(ctx context.Context, dir, appID string, remotes []*models.AppAppBranch) (*models.AppAppBranch, error) {
	name, err := promptWithDefault("App branch name", migrationDefaultBranchName)
	if err != nil {
		return nil, err
	}
	fileName := sanitizeBranchFileName(name)
	if fileName == "" {
		return nil, &ui.CLIUserError{Msg: fmt.Sprintf("app branch name %q has no characters usable in a file name", name)}
	}

	if existing := findBranchByName(remotes, name); existing != nil {
		if existing.ManagedBy != appBranchManagedByConfig {
			return nil, &ui.CLIUserError{Msg: fmt.Sprintf(
				"app branch %q already exists and is managed manually, so a branch file cannot take it over; rerun and pick a different name",
				name,
			)}
		}
		reuse, err := bubbles.InlineConfirm(fmt.Sprintf("App branch %q already exists and is config-managed. Reuse it as is?", name), true, true)
		if err != nil {
			return nil, errMigrationCancelled
		}
		if reuse {
			return existing, nil
		}
	}

	path := branchConfigFilePath(dir, name)
	write := true
	if _, statErr := os.Stat(path); statErr == nil {
		choice, err := chooseExistingFileAction(path)
		if err != nil {
			return nil, err
		}
		switch choice {
		case fileChoiceCancel:
			return nil, errMigrationCancelled
		case fileChoiceUseExisting:
			write = false
		}
	} else if !os.IsNotExist(statErr) {
		return nil, fmt.Errorf("unable to check %s: %w", path, statErr)
	}

	if write {
		cfg, err := s.promptMigrationBranchConfig(ctx, dir, appID, name, remotes)
		if err != nil {
			return nil, err
		}
		if err := writeBranchConfigFile(path, cfg); err != nil {
			return nil, err
		}
		ui.PrintSuccess("wrote " + path)
	}

	branchName := name
	if !write {
		existingCfg, err := parse.ParseAppBranchConfigFile(path)
		if err != nil {
			return nil, err
		}
		branchName = existingCfg.Name
	}

	if err := s.SyncBranches(ctx, SyncBranchesOptions{
		Path:    path,
		AppID:   appID,
		Confirm: true,
	}); err != nil {
		return nil, printedErr{err}
	}

	refreshed, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to list app branches: %w", err)
	}
	branch := findBranchByName(refreshed, branchName)
	if branch == nil {
		return nil, fmt.Errorf("app branch %q was not found after syncing %s", branchName, path)
	}
	return branch, nil
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

func (s *Service) promptMigrationBranchConfig(ctx context.Context, dir, appID, name string, remotes []*models.AppAppBranch) (*config.AppBranchConfig, error) {
	known := s.knownConnectedRepos(ctx, appID, remotes)

	repo, err := selectOrType(
		"Connected GitHub repo (owner/name)",
		repoCandidates(gitOriginRepo(ctx, dir), known),
		"",
	)
	if err != nil {
		return nil, err
	}

	directory, err := promptWithDefault("Directory of the app config within the repo", migrationDefaultDirectory)
	if err != nil {
		return nil, err
	}

	gitBranch, err := selectOrType(
		"Git branch to track",
		gitBranchCandidates(repo, gitCurrentBranch(ctx, dir), known),
		"",
	)
	if err != nil {
		return nil, err
	}

	return newMigrationBranchConfig(name, repo, directory, gitBranch), nil
}

// knownConnectedRepos reads the connected repos the app's branches already
// track; failures only cost the selector a suggestion.
func (s *Service) knownConnectedRepos(ctx context.Context, appID string, remotes []*models.AppAppBranch) []config.ConnectedRepoConfig {
	var out []config.ConnectedRepoConfig
	for _, b := range remotes {
		if b == nil {
			continue
		}
		latest, err := s.latestBranchConfig(ctx, appID, b.ID)
		if err != nil || latest == nil || latest.ConnectedGithubVcsConfig == nil {
			continue
		}
		vcs := latest.ConnectedGithubVcsConfig
		out = append(out, config.ConnectedRepoConfig{Repo: vcs.Repo, Directory: vcs.Directory, Branch: vcs.Branch})
	}
	return out
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

var unsafeFileNameChars = regexp.MustCompile(`[^a-z0-9._-]+`)

func sanitizeBranchFileName(name string) string {
	out := unsafeFileNameChars.ReplaceAllString(strings.ToLower(strings.TrimSpace(name)), "-")
	return strings.Trim(out, "-._")
}

func branchConfigFilePath(dir, name string) string {
	return filepath.Join(dir, migrationBranchesDir, sanitizeBranchFileName(name)+".toml")
}

func promptWithDefault(prompt, def string) (string, error) {
	value, err := bubbles.PromptText(fmt.Sprintf("%s (default %q)", prompt, def), def, "", false, true)
	if err != nil {
		return "", errMigrationCancelled
	}
	if value == "" {
		return def, nil
	}
	return value, nil
}

// selectOrType offers the candidates plus a free-text escape hatch, and falls
// straight to a required text prompt when there is nothing to suggest.
func selectOrType(prompt string, candidates []string, placeholder string) (string, error) {
	if len(candidates) > 0 {
		items := make([]bubbles.SelectorItem, 0, len(candidates)+1)
		for _, c := range candidates {
			items = append(items, bubbles.NewSelectorItem(c, "", c))
		}
		items = append(items, bubbles.NewSelectorItem("Enter a different value", "", migrationEnterOtherValue))
		choice, err := bubbles.SelectFromItems(prompt, items, true)
		if err != nil {
			return "", errMigrationCancelled
		}
		if choice != migrationEnterOtherValue {
			return choice, nil
		}
	}

	value, err := bubbles.PromptText(prompt, placeholder, "", true, true)
	if err != nil {
		return "", errMigrationCancelled
	}
	return value, nil
}

func repoCandidates(gitOrigin string, known []config.ConnectedRepoConfig) []string {
	var out []string
	if gitOrigin != "" {
		out = appendUnique(out, gitOrigin)
	}
	for _, k := range known {
		if k.Repo != "" {
			out = appendUnique(out, k.Repo)
		}
	}
	return out
}

func gitBranchCandidates(repo, current string, known []config.ConnectedRepoConfig) []string {
	var out []string
	if current != "" && current != "HEAD" {
		out = appendUnique(out, current)
	}
	for _, k := range known {
		if k.Branch != "" && strings.EqualFold(k.Repo, repo) {
			out = appendUnique(out, k.Branch)
		}
	}
	return out
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
