package apps

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
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

	remotes, err := s.api.GetAppBranches(ctx, appID)
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
		ok, err := bubbles.ShowConfirmDialog(fmt.Sprintf("apply %d changes to app %s?", changeCount, appID), s.cfg.Interactive)
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
	for _, item := range plan {
		switch item.Op {
		case branchOpCreate:
			ui.PrintRaw(bubbles.Green(fmt.Sprintf("+ create   %s\n", item.Name)))
			ui.PrintRaw(branchDiffToString(item.Diff, "    "))
		case branchOpUpdate:
			ui.PrintRaw(bubbles.Yellow(fmt.Sprintf("~ update   %s\n", item.Name)))
			ui.PrintRaw(branchDiffToString(item.Diff, "    "))
		case branchOpDelete:
			ui.PrintRaw(bubbles.Red(fmt.Sprintf("- delete   %s\n", item.Name)))
			if item.ManagedBy != "" {
				ui.PrintRaw(bubbles.Red(fmt.Sprintf("    managed_by   %s\n", item.ManagedBy)))
			}
		default:
			ui.PrintRaw(fmt.Sprintf("  unchanged  %s\n", item.Name))
		}
	}

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

func branchDiffToString(d *diff.Diff, indent string) string {
	if d == nil {
		return ""
	}
	changed := d.FormatChanged("")
	if changed == "" {
		return ""
	}
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimSuffix(changed, "\n"), "\n") {
		if line == "" {
			continue
		}
		colored := line
		switch {
		case strings.HasPrefix(strings.TrimLeft(line, " \t"), "+"):
			colored = bubbles.Green(line)
		case strings.HasPrefix(strings.TrimLeft(line, " \t"), "-"):
			colored = bubbles.Red(line)
		case strings.HasPrefix(strings.TrimLeft(line, " \t"), "~"):
			colored = bubbles.Yellow(line)
		}
		b.WriteString(indent)
		b.WriteString(colored)
		b.WriteByte('\n')
	}
	return b.String()
}
