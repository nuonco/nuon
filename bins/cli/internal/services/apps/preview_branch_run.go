package apps

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	previewui "github.com/nuonco/nuon/bins/cli/internal/ui/v3/preview"
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

	if s.cfg.Interactive && !asJSON {
		return s.previewBranchRunInteractive(ctx, appID, branchID, opts)
	}

	if branchID == "" {
		return view.Error(fmt.Errorf("app branch required: use --branch-id or run interactively"))
	}
	branchID, err = s.selectBranchID(ctx, appID, branchID)
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

	previewReq, err := s.resolvePreviewRunRequest(ctx, appID, branchID, configID, opts)
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

func (s *Service) previewBranchRunInteractive(ctx context.Context, appID, branchID string, opts PreviewBranchRunOptions) error {
	if opts.PRNumber != nil && opts.GitRef != "" {
		return fmt.Errorf("specify either --pr-number or --git-ref, not both")
	}
	mode, err := parsePreviewMode(opts.Mode)
	if err != nil {
		return err
	}

	branches := make([]previewui.Branch, 0)
	if branchID != "" {
		resolved, err := s.resolveAppBranchID(ctx, appID, branchID)
		if err != nil {
			return err
		}
		branchID = resolved
	} else {
		appBranches, err := s.api.GetAppBranches(ctx, appID)
		if err != nil {
			return fmt.Errorf("unable to list app branches: %w", err)
		}
		if len(appBranches) == 0 {
			return fmt.Errorf("no branches found for this app; create one with: nuon branches create")
		}
		for _, branch := range appBranches {
			branches = append(branches, previewui.Branch{ID: branch.ID, Name: branch.Name})
		}
	}

	loadBranch := func(ctx context.Context, selectedBranchID string) (*previewui.Data, error) {
		branch, err := s.api.GetAppBranch(ctx, appID, selectedBranchID)
		if err != nil {
			return nil, fmt.Errorf("unable to load app branch: %w", err)
		}
		latest, err := s.api.GetAppBranchLatestConfig(ctx, appID, selectedBranchID)
		if err != nil {
			return nil, fmt.Errorf("unable to load branch config: %w", err)
		}
		sources, err := s.api.GetAppBranchPreviewSources(ctx, appID, selectedBranchID)
		if err != nil {
			return nil, fmt.Errorf("unable to list preview sources: %w", err)
		}
		configID := opts.ConfigID
		if configID == "" {
			configID = latest.ID
		}
		candidates, err := s.api.GetAppBranchPreviewInstallCandidates(ctx, appID, selectedBranchID, configID)
		if err != nil {
			return nil, fmt.Errorf("unable to list preview install candidates: %w", err)
		}
		return &previewui.Data{
			BranchName:    branch.Name,
			ConfigID:      configID,
			PreviewConfig: latest.PreviewConfig,
			Sources:       sources,
			Installs:      sortPreviewInstallCandidates(candidates.Installs, selectedBranchID),
		}, nil
	}

	result, err := previewui.App(ctx, branches, loadBranch, previewui.Options{
		BranchID:   branchID,
		Mode:       mode,
		PRNumber:   opts.PRNumber,
		GitRef:     opts.GitRef,
		HeadSHA:    opts.HeadSHA,
		InstallID:  opts.InstallID,
		CurrentRef: currentGitRef(ctx),
	})
	if err != nil {
		return err
	}

	run, err := s.api.TriggerAppBranchRun(ctx, appID, result.BranchID, &models.ServiceTriggerAppBranchRunRequest{
		ConfigID:    result.ConfigID,
		Force:       opts.Force,
		AutoApprove: opts.AutoApprove,
		PreviewRun:  result.Request,
	})
	if err != nil {
		return err
	}

	if opts.Wait && !opts.NoWait && run.WorkflowID != "" {
		if err := s.waitForWorkflowComplete(ctx, run.WorkflowID, false); err != nil {
			return err
		}
		fmt.Printf("Preview run %s completed\n", run.ID)
		return nil
	}
	if opts.NoWait {
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
	appID, branchID, configID string,
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
		if req.HeadSha == "" {
			for _, pr := range sources.PullRequests {
				if pr.PrNumber == req.PrNumber {
					req.HeadSha = pr.HeadSha
					break
				}
			}
		}
	} else if hasBranch {
		req.Source = models.AppAppBranchRunPreviewSourceBranch
		req.GitRef = opts.GitRef
		req.HeadSha = opts.HeadSHA
		if req.HeadSha == "" {
			for _, branch := range sources.Branches {
				if branch.Name == req.GitRef {
					req.HeadSha = branch.Sha
					break
				}
			}
		}
	} else {
		return nil, fmt.Errorf("preview source required: use --pr-number or --git-ref, or run interactively")
	}

	if mode != models.AppAppBranchRunPreviewModeBuildDashOnly {
		req.InstallID = opts.InstallID
	}

	return req, nil
}

func (s *Service) resolvePreviewMode(ctx context.Context, flagMode, configID, appID, branchID string) (models.AppAppBranchRunPreviewMode, error) {
	if flagMode != "" {
		return parsePreviewMode(flagMode)
	}

	if configID != "" {
		cfg, err := s.api.GetAppBranchLatestConfig(ctx, appID, branchID)
		if err == nil && cfg.PreviewConfig != nil && cfg.PreviewConfig.Mode != "" {
			return cfg.PreviewConfig.Mode, nil
		}
	}

	return models.AppAppBranchRunPreviewModePlanDashOnly, nil
}

func parsePreviewMode(mode string) (models.AppAppBranchRunPreviewMode, error) {
	switch mode {
	case "":
		return "", nil
	case "plan-only":
		return models.AppAppBranchRunPreviewModePlanDashOnly, nil
	case "apply":
		return models.AppAppBranchRunPreviewModeApply, nil
	case "build-only":
		return models.AppAppBranchRunPreviewModeBuildDashOnly, nil
	default:
		return "", fmt.Errorf("invalid mode %q: use plan-only, apply, or build-only", mode)
	}
}

func currentGitRef(ctx context.Context) string {
	out, err := exec.CommandContext(ctx, "git", "symbolic-ref", "--quiet", "--short", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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
