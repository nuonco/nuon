package apps

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/diff"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	appBranchManagedByConfig   = "config"
	appBranchManagedByManually = "manually"
)

type SyncBranchesOptions struct {
	Path      string
	AppID     string
	Confirm   bool
	DryRun    bool
	PrintJSON bool
}

type BranchSyncResult struct {
	AppID    string            `json:"app_id"`
	Mode     string            `json:"mode"`
	Applied  bool              `json:"applied"`
	DryRun   bool              `json:"dry_run,omitempty"`
	Summary  BranchSyncSummary `json:"summary"`
	Branches []BranchSyncItem  `json:"branches"`
	Error    string            `json:"error,omitempty"`
}

type BranchSyncSummary struct {
	Created   int `json:"created"`
	Updated   int `json:"updated"`
	Deleted   int `json:"deleted"`
	Unchanged int `json:"unchanged"`
	Failed    int `json:"failed,omitempty"`
}

type BranchSyncItem struct {
	Name      string     `json:"name"`
	Op        string     `json:"op"`
	Status    string     `json:"status"`
	BranchID  string     `json:"branch_id,omitempty"`
	ConfigID  string     `json:"config_id,omitempty"`
	ManagedBy string     `json:"managed_by,omitempty"`
	Diff      *diff.Diff `json:"diff,omitempty"`
	Error     string     `json:"error,omitempty"`
}

func (s *Service) SyncBranches(ctx context.Context, opts SyncBranchesOptions) error {
	if opts.Path == "" {
		return ui.PrintError(&ui.CLIUserError{Msg: "--file is required"})
	}

	appID, err := s.resolveAppID(ctx, opts.AppID)
	if err != nil {
		return ui.PrintError(err)
	}

	files, directory, err := parse.LoadAppBranchConfigs(opts.Path)
	if err != nil {
		return ui.PrintError(err)
	}

	if !opts.PrintJSON {
		if directory {
			ui.PrintLn(fmt.Sprintf("loading branch configs from %s", opts.Path))
		} else {
			ui.PrintLn("single-file mode: only this branch is reconciled; no branches will be deleted")
		}
		for _, f := range files {
			ui.PrintLn(fmt.Sprintf("  %s → %s", f.Path, f.Config.Name))
		}
	}

	remotes, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(fmt.Errorf("unable to list branches for app %s: %w", appID, err))
	}

	resolver := newBranchNameResolver(s.api, appID)
	remoteByName := make(map[string]*models.AppAppBranch, len(remotes))
	remoteCfg := make(map[string]*config.AppBranchConfig, len(remotes))
	for _, remote := range remotes {
		remoteByName[remote.Name] = remote
		latest, err := s.latestBranchConfig(ctx, appID, remote.ID)
		if err != nil {
			return ui.PrintError(err)
		}
		normalized, err := normalizeRemoteBranch(ctx, resolver, remote.Name, latest)
		if err != nil {
			return ui.PrintError(err)
		}
		remoteCfg[remote.Name] = normalized
	}

	local := make([]*config.AppBranchConfig, 0, len(files))
	for _, f := range files {
		canonical, err := canonicalizeLocalBranch(ctx, resolver, f.Config)
		if err != nil {
			return ui.PrintError(err)
		}
		local = append(local, canonical)
	}

	plan, err := buildBranchSyncPlan(local, remotes, remoteCfg, directory)
	if err != nil {
		return ui.PrintError(err)
	}

	result := newBranchSyncResult(appID, directory, plan)
	result.DryRun = opts.DryRun

	if !opts.PrintJSON {
		printBranchPlan(plan)
	}

	changeCount := result.Summary.Created + result.Summary.Updated + result.Summary.Deleted
	if opts.DryRun || changeCount == 0 {
		result.Applied = false
		return printBranchSyncResult(result, opts.PrintJSON, opts.DryRun)
	}

	if !opts.Confirm {
		if !s.cfg.Interactive {
			return ui.PrintError(&ui.CLIUserError{Msg: "use --confirm to apply"})
		}
		ok, err := bubbles.ShowConfirmDialog(s.branchSyncConfirmPrompt(ctx, changeCount, appID), s.cfg.Interactive)
		if err != nil {
			return ui.PrintError(err)
		}
		if !ok {
			ui.PrintLn("sync aborted")
			result.Applied = false
			return printBranchSyncResult(result, opts.PrintJSON, false)
		}
	}

	if err := s.applyBranchPlan(ctx, appID, plan, &result); err != nil {
		_ = printBranchSyncResult(result, opts.PrintJSON, false)
		return ui.PrintError(err)
	}

	result.Applied = true
	return printBranchSyncResult(result, opts.PrintJSON, false)
}

func (s *Service) branchSyncConfirmPrompt(ctx context.Context, changeCount int, appID string) string {
	if app, err := s.api.GetApp(ctx, appID); err == nil && app != nil && app.Name != "" {
		return fmt.Sprintf("apply %d changes to app %s (%s)?", changeCount, app.Name, appID)
	}
	return fmt.Sprintf("apply %d changes to app %s?", changeCount, appID)
}

func (s *Service) latestBranchConfig(ctx context.Context, appID, branchID string) (*models.AppAppBranchConfig, error) {
	cfg, err := s.api.GetAppBranchLatestConfig(ctx, appID, branchID)
	if err != nil {
		if nuon.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("unable to get latest config for branch %s: %w", branchID, err)
	}
	return cfg, nil
}

func newBranchSyncResult(appID string, directory bool, plan []branchPlanItem) BranchSyncResult {
	mode := "file"
	if directory {
		mode = "directory"
	}
	result := BranchSyncResult{
		AppID:    appID,
		Mode:     mode,
		Branches: make([]BranchSyncItem, 0, len(plan)),
	}
	for _, item := range plan {
		result.Branches = append(result.Branches, BranchSyncItem{
			Name:      item.Name,
			Op:        string(item.Op),
			Status:    "planned",
			BranchID:  item.BranchID,
			ManagedBy: item.ManagedBy,
			Diff:      item.Diff,
		})
		switch item.Op {
		case branchOpCreate:
			result.Summary.Created++
		case branchOpUpdate:
			result.Summary.Updated++
		case branchOpDelete:
			result.Summary.Deleted++
		default:
			result.Summary.Unchanged++
		}
	}
	return result
}

func printBranchSyncResult(result BranchSyncResult, asJSON, dryRun bool) error {
	if asJSON {
		ui.PrintJSON(result)
		return nil
	}
	if dryRun {
		ui.PrintLn("dry run: no changes applied")
		return nil
	}
	ui.PrintSuccess(fmt.Sprintf(
		"synced branches (%d created, %d updated, %d deleted, %d unchanged)",
		result.Summary.Created, result.Summary.Updated, result.Summary.Deleted, result.Summary.Unchanged,
	))
	return nil
}

func printBranchPlan(plan []branchPlanItem) {
	ui.PrintLn("[branch plan]")
	ui.PrintRaw(formatBranchPlan(plan))

	var created, updated, deleted, unchanged int
	for _, item := range plan {
		switch item.Op {
		case branchOpCreate:
			created++
		case branchOpUpdate:
			updated++
		case branchOpDelete:
			deleted++
		default:
			unchanged++
		}
	}
	ui.PrintLn(fmt.Sprintf("(create %d, update %d, delete %d, unchanged %d)", created, updated, deleted, unchanged))
}

const branchHeaderMinWidth = 40

// formatBranchPlan renders one header row per branch, then that branch's diff
// set off by a left-hand rail. A blank line under each diff separates branches.
func formatBranchPlan(plan []branchPlanItem) string {
	nameWidth := 0
	for _, item := range plan {
		if n := utf8.RuneCountInString(item.Name); n > nameWidth {
			nameWidth = n
		}
	}
	rule := branchHeaderRule(nameWidth)

	var b strings.Builder
	for _, item := range plan {
		changed := item.Op != branchOpUnchanged
		b.WriteString(formatBranchHeader(item.Name, nameWidth, changed))
		b.WriteByte('\n')
		b.WriteString(rule)
		b.WriteByte('\n')
		if !changed {
			continue
		}
		switch item.Op {
		case branchOpCreate:
			writeBranchDiffLine(&b, "+ create", '+')
			writeBranchDiffBody(&b, item.Diff)
		case branchOpUpdate:
			writeBranchDiffLine(&b, "~ update", '~')
			writeBranchDiffBody(&b, item.Diff)
		case branchOpDelete:
			writeBranchDiffLine(&b, "- delete", '-')
			if item.ManagedBy != "" {
				writeBranchDiffLine(&b, fmt.Sprintf("    managed_by   %s", item.ManagedBy), '-')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func formatBranchHeader(name string, nameWidth int, changed bool) string {
	status := "unchanged"
	statusText := styles.TextSubtle.Render(status)
	if changed {
		status = "changed"
		statusText = styles.TextWarning.Render(status)
	}
	bar := styles.TextSubtle.Render("|")
	nameCell := styles.TextBold.Render(name) + strings.Repeat(" ", nameWidth-utf8.RuneCountInString(name))
	return fmt.Sprintf("%s name: %s %s %s", bar, nameCell, bar, statusText)
}

func branchHeaderRule(nameWidth int) string {
	width := utf8.RuneCountInString("| name: ") + nameWidth + utf8.RuneCountInString(" | ") + len("unchanged")
	if width < branchHeaderMinWidth {
		width = branchHeaderMinWidth
	}
	return styles.TextSubtle.Render(strings.Repeat("─", width))
}

func writeBranchDiffBody(b *strings.Builder, d *diff.Diff) {
	if d == nil {
		return
	}
	changed := d.FormatChanged("")
	if changed == "" {
		return
	}
	for _, line := range strings.Split(strings.TrimSuffix(changed, "\n"), "\n") {
		if line == "" {
			continue
		}
		// Nest the field diff under the operation line. FormatChanged puts the
		// +/-/~ marker before its own indent, so keep a fixed gutter here.
		writeBranchDiffLine(b, "    "+line, 0)
	}
}

func writeBranchDiffLine(b *strings.Builder, plain string, kind rune) {
	plain = strings.ReplaceAll(plain, "\t", "  ")
	if kind == 0 {
		kind = diffLineKind(plain)
	}
	gutter := styles.TextSubtle.Render("│")
	content := plain
	switch kind {
	case '+':
		gutter = styles.TextSuccess.Render("│")
		content = styles.TextSuccess.Render(plain)
	case '-':
		gutter = styles.TextError.Render("│")
		content = styles.TextError.Render(plain)
	case '~':
		gutter = styles.TextWarning.Render("│")
		content = styles.TextWarning.Render(plain)
	}
	b.WriteString(gutter)
	b.WriteByte(' ')
	b.WriteString(content)
	b.WriteByte('\n')
}

func diffLineKind(line string) rune {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return 0
	}
	switch trimmed[0] {
	case '+', '-', '~':
		return rune(trimmed[0])
	default:
		return 0
	}
}
