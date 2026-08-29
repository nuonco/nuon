package apps

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type PreviewBranchRunOptions struct {
	ConfigID    string
	PRNumber    *int
	GitRef      string
	HeadSHA     string
	InstallID   string
	Mode        string
	Force       bool
	AutoApprove bool
	Wait        bool
	NoWait      bool
}

func (s *Service) PreviewBranchRun(ctx context.Context, appID, branchID string, opts PreviewBranchRunOptions, asJSON bool) error {
	view := ui.NewGetView()

	appID, err := s.resolveAppID(ctx, appID)
	if err != nil {
		return view.Error(err)
	}

	branchID, err = s.selectBranchID(ctx, appID, branchID)
	if err != nil {
		return view.Error(err)
	}

	branch, err := s.api.GetAppBranch(ctx, appID, branchID)
	if err != nil {
		return view.Error(err)
	}

	configID := opts.ConfigID
	if configID == "" {
		latest, err := s.api.GetAppBranchLatestConfig(ctx, appID, branchID)
		if err != nil {
			return view.Error(fmt.Errorf("unable to load branch config: %w", err))
		}
		configID = latest.ID
	}

	previewReq, err := s.resolvePreviewRunRequest(ctx, appID, branchID, branch.Name, configID, opts)
	if err != nil {
		return view.Error(err)
	}

	triggerReq := &models.ServiceTriggerAppBranchRunRequest{
		ConfigID:    configID,
		Force:       opts.Force,
		AutoApprove: opts.AutoApprove,
		PreviewRun:  previewReq,
	}

	run, err := s.api.TriggerAppBranchRun(ctx, appID, branchID, triggerReq)
	if err != nil {
		return view.Error(err)
	}

	if opts.Wait && !opts.NoWait && run.WorkflowID != "" {
		if err := s.waitForWorkflowComplete(ctx, run.WorkflowID, asJSON); err != nil {
			return view.Error(err)
		}
		if asJSON {
			ui.PrintJSON(run)
			return nil
		}
		fmt.Printf("Preview run %s completed\n", run.ID)
		return nil
	}

	if asJSON || opts.NoWait {
		ui.PrintJSON(run)
		return nil
	}

	if run.WorkflowID == "" {
		fmt.Printf("Triggered preview run %s\n", run.ID)
		return nil
	}

	workflow.WorkflowApp(ctx, s.cfg, s.api, "", run.WorkflowID, false)
	return nil
}

func (s *Service) resolvePreviewRunRequest(
	ctx context.Context,
	appID, branchID, branchName, configID string,
	opts PreviewBranchRunOptions,
) (*models.ServicePreviewRunRequest, error) {
	mode, err := s.resolvePreviewMode(ctx, opts.Mode, configID, appID, branchID)
	if err != nil {
		return nil, err
	}

	if opts.PRNumber != nil && opts.GitRef != "" {
		return nil, fmt.Errorf("specify either --pr-number or --git-ref, not both")
	}

	hasPR := opts.PRNumber != nil
	hasBranch := opts.GitRef != ""
	needsInteractiveSource := !hasPR && !hasBranch

	sources, err := s.api.GetAppBranchPreviewSources(ctx, appID, branchID)
	if err != nil {
		return nil, fmt.Errorf("unable to list preview sources: %w", err)
	}

	req := &models.ServicePreviewRunRequest{
		Mode: mode,
	}

	if hasPR {
		req.Source = models.AppAppBranchRunPreviewSourcePr
		req.PrNumber = int64(*opts.PRNumber)
		req.HeadSha = opts.HeadSHA
	} else if hasBranch {
		req.Source = models.AppAppBranchRunPreviewSourceBranch
		req.GitRef = opts.GitRef
		req.HeadSha = opts.HeadSHA
	} else if needsInteractiveSource {
		if !s.cfg.Interactive {
			return nil, fmt.Errorf("preview source required: use --pr-number or --git-ref, or run interactively")
		}
		if err := s.fillInteractiveSource(req, sources, opts.HeadSHA); err != nil {
			return nil, err
		}
	}

	if mode != models.AppAppBranchRunPreviewModeBuildDashOnly {
		installID, err := s.resolvePreviewInstallID(ctx, appID, branchID, configID, branchName, opts.InstallID)
		if err != nil {
			return nil, err
		}
		req.InstallID = installID
	}

	return req, nil
}

func (s *Service) resolvePreviewMode(ctx context.Context, flagMode, configID, appID, branchID string) (models.AppAppBranchRunPreviewMode, error) {
	if flagMode != "" {
		switch flagMode {
		case "plan-only":
			return models.AppAppBranchRunPreviewModePlanDashOnly, nil
		case "apply":
			return models.AppAppBranchRunPreviewModeApply, nil
		case "build-only":
			return models.AppAppBranchRunPreviewModeBuildDashOnly, nil
		default:
			return "", fmt.Errorf("invalid mode %q: use plan-only, apply, or build-only", flagMode)
		}
	}

	if configID != "" {
		cfg, err := s.api.GetAppBranchLatestConfig(ctx, appID, branchID)
		if err == nil && cfg.PreviewConfig != nil && cfg.PreviewConfig.Mode != "" {
			return cfg.PreviewConfig.Mode, nil
		}
	}

	if s.cfg.Interactive {
		selected, err := bubbles.SelectFromItems("Select preview mode", []bubbles.SelectorItem{
			bubbles.NewSelectorItem("Plan only", "", string(models.AppAppBranchRunPreviewModePlanDashOnly)),
			bubbles.NewSelectorItem("Apply", "", string(models.AppAppBranchRunPreviewModeApply)),
			bubbles.NewSelectorItem("Build only", "", string(models.AppAppBranchRunPreviewModeBuildDashOnly)),
		}, true)
		if err != nil {
			return "", err
		}
		return models.AppAppBranchRunPreviewMode(selected), nil
	}

	return models.AppAppBranchRunPreviewModePlanDashOnly, nil
}

func (s *Service) fillInteractiveSource(
	req *models.ServicePreviewRunRequest,
	sources *models.HelpersListPreviewSourcesResult,
	headSHA string,
) error {
	sourceKind, err := bubbles.SelectFromItems("Select preview source", []bubbles.SelectorItem{
		bubbles.NewSelectorItem("Pull request", "", "pr"),
		bubbles.NewSelectorItem("Git branch", "", "branch"),
	}, true)
	if err != nil {
		return err
	}

	switch sourceKind {
	case "pr":
		if len(sources.PullRequests) == 0 {
			return fmt.Errorf("no open pull requests found for this branch")
		}
		items := make([]bubbles.SelectorItem, len(sources.PullRequests))
		for i, pr := range sources.PullRequests {
			items[i] = bubbles.NewSelectorItem(
				fmt.Sprintf("#%d · %s", pr.PrNumber, pr.Title),
				pr.HeadRef,
				strconv.FormatInt(pr.PrNumber, 10),
			)
		}
		selected, err := bubbles.SelectFromItems("Select a pull request", items, true)
		if err != nil {
			return err
		}
		prNum, err := strconv.ParseInt(selected, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid pull request number: %w", err)
		}
		req.Source = models.AppAppBranchRunPreviewSourcePr
		req.PrNumber = prNum
		for _, pr := range sources.PullRequests {
			if pr.PrNumber == prNum {
				if headSHA != "" {
					req.HeadSha = headSHA
				} else {
					req.HeadSha = pr.HeadSha
				}
				break
			}
		}
	case "branch":
		if len(sources.Branches) == 0 {
			return fmt.Errorf("no git branches available for preview")
		}
		items := make([]bubbles.SelectorItem, len(sources.Branches))
		for i, b := range sources.Branches {
			items[i] = bubbles.NewSelectorItem(b.Name, "", b.Name)
		}
		selected, err := bubbles.SelectFromItems("Select a git branch", items, true)
		if err != nil {
			return err
		}
		req.Source = models.AppAppBranchRunPreviewSourceBranch
		req.GitRef = selected
		for _, b := range sources.Branches {
			if b.Name == selected {
				if headSHA != "" {
					req.HeadSha = headSHA
				} else {
					req.HeadSha = b.Sha
				}
				break
			}
		}
	default:
		return fmt.Errorf("unknown preview source %q", sourceKind)
	}

	return nil
}

func (s *Service) resolvePreviewInstallID(
	ctx context.Context,
	appID, branchID, configID, branchName, flagInstallID string,
) (string, error) {
	if flagInstallID != "" {
		return flagInstallID, nil
	}

	if !s.cfg.Interactive {
		return "", fmt.Errorf("install required for this preview mode: use --install-id or run interactively")
	}

	candidates, err := s.api.GetAppBranchPreviewInstallCandidates(ctx, appID, branchID, configID)
	if err != nil {
		return "", fmt.Errorf("unable to list preview install candidates: %w", err)
	}
	if len(candidates.Installs) == 0 {
		return "", fmt.Errorf("no installs found for this app")
	}

	sorted := sortPreviewInstallCandidates(candidates.Installs, branchID)
	options := make([]bubbles.InstallOption, len(sorted))
	for i, inst := range sorted {
		opt := bubbles.InstallOption{ID: inst.ID, Name: inst.Name}
		if inst.AppBranchID != "" && inst.AppBranchID != branchID {
			branchLabel := branchName
			if inst.AppBranch != nil && inst.AppBranch.Name != "" {
				branchLabel = inst.AppBranch.Name
			}
			opt.Description = fmt.Sprintf("On branch %s", branchLabel)
		}
		options[i] = opt
	}

	return bubbles.SelectInstall(options, true)
}

func sortPreviewInstallCandidates(installs []*models.AppInstall, branchID string) []*models.AppInstall {
	sorted := append([]*models.AppInstall(nil), installs...)
	rank := func(inst *models.AppInstall) int {
		if inst.AppBranchID == branchID {
			return 0
		}
		if inst.AppBranchID == "" {
			return 1
		}
		return 2
	}

	sort.SliceStable(sorted, func(i, j int) bool {
		ri, rj := rank(sorted[i]), rank(sorted[j])
		if ri != rj {
			return ri < rj
		}
		if ri == 2 {
			bi, bj := "", ""
			if sorted[i].AppBranch != nil {
				bi = sorted[i].AppBranch.Name
			}
			if sorted[j].AppBranch != nil {
				bj = sorted[j].AppBranch.Name
			}
			if bi != bj {
				return bi < bj
			}
		}
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	return sorted
}
