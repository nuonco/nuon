package installs

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/pkg/browser"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/installcreate"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/services/labels"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	appselector "github.com/nuonco/nuon/bins/cli/internal/ui/v3/app/selector"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/install/creator"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
)

const (
	statusError       = "error"
	statusActive      = "active"
	statusAccessError = "access_error"

	skipBranchID    = "\x00skip-branch"
	skipBranchLabel = "Skip app branch (use the most recent config from `nuon apps sync`)"
	skipGroupID     = "\x00skip-group"
	skipGroupLabel  = "Skip install group (the install will be orphaned until its labels match)"
)

// TargetAccount is the cloud account an install is pinned to. Only the field matching the
// app's cloud platform is used, and the API requires it once the org has phone-home auth
// enabled.
type TargetAccount struct {
	AWSAccountID        string
	AzureSubscriptionID string
	GCPProjectID        string
}

func (s *Service) Create(ctx context.Context, appID, name, region string, target TargetAccount, inputs, labelArgs []string, asJSON, noSelect, stackOnly bool, appBranchID, installGroupID string) error {
	if appID == "" {
		selectedID, err := appselector.App(ctx, s.cfg, s.api)
		if err != nil {
			return ui.PrintError(err)
		}
		appID = selectedID
	} else {
		var err error
		appID, err = lookup.AppID(ctx, s.api, appID)
		if err != nil {
			return ui.PrintError(err)
		}
	}

	labelsMap, removeKeys, err := labels.ParseArgs(labelArgs)
	if err != nil {
		return ui.PrintError(err)
	}
	if len(removeKeys) > 0 {
		return ui.PrintError(fmt.Errorf("label removal (key-) is not allowed at install creation; use `nuon installs label` after the install exists"))
	}

	branch, err := s.resolveCreateInstallBranch(ctx, appID, appBranchID, installGroupID, asJSON)
	if err != nil {
		return ui.PrintError(err)
	}
	branchID := branch.branchID
	if branch.groupLabels != nil {
		if err := installcreate.MergeGroupLabels(labelsMap, branch.groupLabels); err != nil {
			return ui.PrintError(err)
		}
	}

	if s.cfg.Preview && !asJSON {
		var inputConfig *models.AppAppInputConfig
		if branchID != "" {
			inputConfig, err = installcreate.ResolveInputConfig(ctx, s.api, appID, branchID)
			if err != nil {
				fmt.Fprintln(os.Stderr, bubbles.ErrorStyle.Render(err.Error()))
				return err
			}
		}
		installID, _ := creator.InstallCreatorApp(
			ctx,
			s.cfg,
			s.api,
			appID,
			name,
			region,
			labelsMap,
			branchID,
			inputConfig,
			branch.groups,
		)
		if installID == "" {
			ui.PrintLn("no install created")
			return nil
		}
		ui.PrintLn(fmt.Sprintf("fetching workflow for new install: %s", installID))
		// get the first workflow for this install and open it
		workflows, _, err := s.api.GetWorkflows(ctx, installID, &models.GetPaginatedQuery{Limit: 1, Offset: 0})
		if err != nil {
			return ui.PrintError(errors.Wrap(err, "failed to get initial workflow for this new install"))
		}
		if len(workflows) == 0 {
			ui.PrintLn(fmt.Sprintf("install %s has no workflow yet", installID))
			return nil
		}
		wf := workflows[0]
		workflow.WorkflowApp(ctx, s.cfg, s.api, installID, wf.ID, false)
		return nil
	}

	// Outside the creator TUI the group picker is a standalone prompt.
	if branch.groupLabels == nil && len(branch.groups) > 0 && !asJSON && s.cfg.Interactive {
		groupLabels, err := promptInstallGroup(branch.groups, s.cfg.Interactive)
		if err != nil {
			return ui.PrintError(err)
		}
		if err := installcreate.MergeGroupLabels(labelsMap, groupLabels); err != nil {
			return ui.PrintError(err)
		}
	}

	// we collect these and pass them down so we can pre-fill specific fields
	inputsMap, err := parseInstallInputs(inputs)
	if err != nil {
		return ui.PrintError(err)
	}

	req, err := s.buildCreateInstallRequest(ctx, appID, name, region, target, inputsMap, labelsMap, branchID)
	if err != nil {
		return ui.PrintError(err)
	}
	req.StackOnly = stackOnly

	if asJSON {
		install, err := s.api.CreateInstall(ctx, appID, req)
		if err != nil {
			return ui.PrintJSONError(err)
		}
		ui.PrintJSON(install)
		return nil
	}

	install, err := s.api.CreateInstall(ctx, appID, req)
	if err != nil {
		return ui.PrintError(fmt.Errorf("error creating install: %w", err))
	}

	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	ui.PrintLn(fmt.Sprintf("install ID: %s", install.ID))

	url := fmt.Sprintf("%s/%s/installs/%s", cfg.DashboardURL, s.cfg.OrgID, install.ID)
	browser.OpenURL(url)

	return nil
}

func (s *Service) buildCreateInstallRequest(ctx context.Context, appID, name, region string, target TargetAccount, inputs, labelsMap map[string]string, appBranchID string) (*models.ServiceCreateInstallRequest, error) {
	inputsWithDefaults, err := s.inputsWithDefaults(ctx, appID, appBranchID, inputs)
	if err != nil {
		return nil, err
	}

	req := &models.ServiceCreateInstallRequest{
		Name:        &name,
		Inputs:      inputsWithDefaults,
		Labels:      labelsMap,
		AppBranchID: appBranchID,
	}

	runnerCfg, err := s.api.GetAppRunnerLatestConfig(ctx, appID)
	if err != nil || runnerCfg == nil {
		req.AwsAccount = &models.HelpersCreateInstallAWSAccountParams{Region: region, AccountID: target.AWSAccountID}
		return req, nil
	}

	switch runnerCfg.CloudPlatform {
	case models.AppCloudPlatformGcp:
		if err := requireInstallRegion(region, "GCP"); err != nil {
			return nil, err
		}
		req.GcpAccount = &models.HelpersCreateInstallGCPAccountParams{
			ProjectID: target.GCPProjectID,
			Region:    region,
		}
	case models.AppCloudPlatformAzure:
		if err := requireInstallRegion(region, "Azure"); err != nil {
			return nil, err
		}
		req.AzureAccount = &models.HelpersCreateInstallAzureAccountParams{
			SubscriptionID: target.AzureSubscriptionID,
			Location:       region,
		}
	default:
		if err := requireInstallRegion(region, "AWS"); err != nil {
			return nil, err
		}
		req.AwsAccount = &models.HelpersCreateInstallAWSAccountParams{Region: region, AccountID: target.AWSAccountID}
	}

	return req, nil
}

func requireInstallRegion(region, cloud string) error {
	if region == "" {
		return fmt.Errorf("--region is required for %s installs", cloud)
	}
	return nil
}

func parseInstallInputs(inputs []string) (map[string]string, error) {
	inputsMap := make(map[string]string, len(inputs))
	for _, input := range inputs {
		name, value, ok := strings.Cut(input, "=")
		if !ok || name == "" {
			return nil, fmt.Errorf("invalid input %q: expected name=value", input)
		}
		inputsMap[name] = value
	}
	return inputsMap, nil
}

// inputsWithDefaults merges app input defaults with any explicitly provided values.
// Explicit values win; defaults fill in anything not provided.
func (s *Service) inputsWithDefaults(ctx context.Context, appID, appBranchID string, provided map[string]string) (map[string]string, error) {
	inputCfg, err := installcreate.ResolveInputConfig(ctx, s.api, appID, appBranchID)
	if err != nil || inputCfg == nil {
		if appBranchID != "" {
			if err == nil {
				err = fmt.Errorf("selected app branch %s has no input config", appBranchID)
			}
			return nil, err
		}
		return provided, nil
	}

	merged := make(map[string]string)
	for _, input := range inputCfg.Inputs {
		if input == nil || input.Name == "" || input.Default == "" {
			continue
		}
		merged[input.Name] = input.Default
	}
	for k, v := range provided {
		merged[k] = v
	}
	return merged, nil
}

// createInstallBranch is the outcome of branch selection. groupLabels is only set
// when the group was already decided (an explicit --install-group-id); otherwise
// groups holds the choices to offer later, after the install form.
type createInstallBranch struct {
	branchID    string
	groups      []*models.AppAppBranchInstallGroup
	groupLabels map[string]string
}

func (s *Service) resolveCreateInstallBranch(ctx context.Context, appID, appBranchID, installGroupID string, asJSON bool) (*createInstallBranch, error) {
	branches, err := nuon.GetAllAppBranches(ctx, s.api, appID)
	if err != nil {
		return nil, err
	}
	if len(branches) == 0 {
		return &createInstallBranch{}, nil
	}

	if appBranchID == "" && installGroupID != "" {
		return nil, fmt.Errorf("--install-group-id requires --app-branch-id")
	}
	if appBranchID == "" && (asJSON || !s.cfg.Interactive) {
		return &createInstallBranch{}, nil
	}

	var selectedBranch *models.AppAppBranch
	if appBranchID != "" {
		for _, branch := range branches {
			if branch != nil && branch.ID == appBranchID {
				selectedBranch = branch
				break
			}
		}
		if selectedBranch == nil {
			return nil, fmt.Errorf("app branch %q was not found on this app", appBranchID)
		}
	} else {
		items := make([]bubbles.SelectorItem, 0, len(branches))
		for _, branch := range branches {
			if branch != nil {
				items = append(items, bubbles.NewSelectorItem(branch.Name, branch.ID, branch.ID))
			}
		}
		items = append(items, bubbles.NewSelectorItem(
			skipBranchLabel,
			"",
			skipBranchID,
		))
		selectedID, err := bubbles.SelectFromItems("Select an app branch", items, s.cfg.Interactive)
		if err != nil {
			return nil, err
		}
		if selectedID == skipBranchID {
			return &createInstallBranch{}, nil
		}
		for _, branch := range branches {
			if branch != nil && branch.ID == selectedID {
				selectedBranch = branch
				break
			}
		}
	}

	var groups []*models.AppAppBranchInstallGroup
	if len(selectedBranch.Configs) > 0 && selectedBranch.Configs[0] != nil {
		groups = selectedBranch.Configs[0].InstallGroups
	}

	eligible := installcreate.EligibleGroups(groups)
	if len(eligible) == 0 {
		if installGroupID != "" {
			return nil, fmt.Errorf("app branch %q has no selectable label-based install groups", selectedBranch.Name)
		}
		return &createInstallBranch{branchID: selectedBranch.ID}, nil
	}

	if installGroupID == "" {
		return &createInstallBranch{branchID: selectedBranch.ID, groups: eligible}, nil
	}

	for _, group := range eligible {
		if group.ID == installGroupID {
			groupLabels, err := installcreate.GroupLabels(group)
			if err != nil {
				return nil, err
			}
			return &createInstallBranch{branchID: selectedBranch.ID, groupLabels: groupLabels}, nil
		}
	}
	return nil, fmt.Errorf("install group %q is not a selectable label-based group on app branch %q", installGroupID, selectedBranch.Name)
}

// promptInstallGroup asks which group a new install should join. A nil result
// means the install is left orphaned until its labels match a group.
func promptInstallGroup(eligible []*models.AppAppBranchInstallGroup, interactive bool) (map[string]string, error) {
	items := make([]bubbles.SelectorItem, 0, len(eligible))
	for _, group := range eligible {
		items = append(items, bubbles.NewSelectorItem(group.Name, group.ID, group.ID))
	}
	items = append(items, bubbles.NewSelectorItem(skipGroupLabel, "", skipGroupID))

	selectedID, err := bubbles.SelectFromItems("Select an install group", items, interactive)
	if err != nil {
		return nil, err
	}
	if selectedID == skipGroupID {
		return nil, nil
	}

	for _, group := range eligible {
		if group.ID == selectedID {
			return installcreate.GroupLabels(group)
		}
	}
	return nil, nil
}
