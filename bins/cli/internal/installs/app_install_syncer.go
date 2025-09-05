package installs

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/browser"
	"github.com/pterm/pterm"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/config"
)

const ManagedByNuonCLIConfig = "nuon/cli/install-config"

const defaultPollDuration = time.Second * 10

type appInstallSyncer struct {
	api          nuon.Client
	appID, orgID string
}

func newAppInstallSyncer(api nuon.Client, appID, orgID string) *appInstallSyncer {
	return &appInstallSyncer{
		api:   api,
		appID: appID,
		orgID: orgID,
	}
}

func (s *appInstallSyncer) syncInstall(
	ctx context.Context, installCfg *config.Install, installID string, autoApprove, wait bool,
) (*models.AppInstall, error) {
	var err error
	ui.PrintLn(fmt.Sprintf("syncing install %s", installCfg.Name))

	if installCfg == nil {
		return nil, fmt.Errorf("install config cannot be nil")
	}

	if installID == "" {
		appInstall, err := s.syncNewInstall(ctx, installCfg, autoApprove, wait)
		return appInstall, err
	}

	appInstall, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("error getting install %s: %w", installCfg.Name, err)
	}

	appInstall, err = s.syncExistingInstall(ctx, installCfg, appInstall, autoApprove, wait)
	return appInstall, err
}

func (s *appInstallSyncer) syncNewInstall(ctx context.Context, installCfg *config.Install, autoApprove, wait bool) (*models.AppInstall, error) {
	// Use defaults for any missing inputs.
	{
		appInputCfg, err := s.api.GetAppInputLatestConfig(ctx, s.appID)
		if err != nil {
			return nil, fmt.Errorf("error getting latest input config for app %s: %w", s.appID, err)
		}

		inputDefaults := make(map[string]string)
		for _, ic := range appInputCfg.Inputs {
			if ic.Default != "" {
				inputDefaults[ic.Name] = ic.Default
			}
		}
		installCfg.InputGroups = append([]config.InputGroup{inputDefaults}, installCfg.InputGroups...)
	}

	diff, _, err := installCfg.Diff(nil)
	pterm.DefaultBasicText.Println(diff)

	if !autoApprove {
		ok, err := pterm.DefaultInteractiveConfirm.Show("Do you want to proceed with creating this install?")
		if err != nil {
			return nil, fmt.Errorf("error getting confirmation: %w", err)
		}
		if !ok {
			ui.PrintSuccess(fmt.Sprintf("skipping install %s, sync aborted by user", installCfg.Name))
			return nil, nil
		}
	}

	req := models.ServiceCreateInstallRequest{
		Name:   &installCfg.Name,
		Inputs: installCfg.FlattenedInputs(),
		Metadata: &models.HelpersInstallMetadata{
			ManagedBy: ManagedByNuonCLIConfig,
		},
	}
	if installCfg.AWSAccount != nil {
		req.AwsAccount = &models.ServiceCreateInstallRequestAwsAccount{
			Region: installCfg.AWSAccount.Region,
		}
	}
	if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown {
		req.InstallConfig = &models.HelpersCreateInstallConfigParams{
			ApprovalOption: installCfg.ApprovalOption.APIType(),
		}
	}

	appInstall, installWorkflowID, err := s.api.CreateInstall(ctx, s.appID, &req)
	if err != nil {
		return nil, fmt.Errorf("error creating install %s: %w", installCfg.Name, err)
	}

	s.printWorkflowStatus(ctx, installWorkflowID, appInstall.ID, wait)

	ui.PrintSuccess(fmt.Sprintf("install %s created successfully", appInstall.Name))
	return appInstall, nil
}

func (s *appInstallSyncer) syncExistingInstall(
	ctx context.Context, installCfg *config.Install, appInstall *models.AppInstall, autoApprove, wait bool,
) (*models.AppInstall, error) {
	var err error

	currInputs, err := s.api.GetInstallCurrentInputs(ctx, appInstall.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting current inputs for install %s: %w", appInstall.Name, err)
	}

	appInputCfg, err := s.api.GetAppInputLatestConfig(ctx, appInstall.AppID)
	if err != nil {
		return nil, fmt.Errorf("error getting current app input config for install %s: %w", appInstall.Name, err)
	}

	upstreamConfig := &config.Install{}
	upstreamConfig.ParseIntoInstall(appInstall, currInputs, appInputCfg, false)

	diff, diffRes, err := installCfg.Diff(upstreamConfig)
	if err != nil {
		return nil, fmt.Errorf("error generating diff for install %s: %w", installCfg.Name, err)
	}
	if !diffRes.HasChanged {
		ui.PrintSuccess(fmt.Sprintf("install %s is up to date, no changes needed", installCfg.Name))
		return appInstall, nil
	}

	pterm.DefaultBasicText.Printf(`[install diff]
%s
(added %d, removed %d, changed %d)
`, diff, diffRes.Added, diffRes.Removed, diffRes.Changed)

	if !autoApprove {
		ok, err := pterm.DefaultInteractiveConfirm.Show("Do you want to proceed with creating this install?")
		if err != nil {
			return nil, fmt.Errorf("error getting confirmation: %w", err)
		}
		if !ok {
			ui.PrintSuccess(fmt.Sprintf("skipping install %s, sync aborted by user", installCfg.Name))
			return nil, nil
		}
	}

	if installCfg.ApprovalOption != config.InstallApprovalOptionUnknown {
		if appInstall.InstallConfig == nil {
			appInstall.InstallConfig, err = s.api.CreateInstallConfig(ctx, appInstall.ID, &models.ServiceCreateInstallConfigRequest{
				ApprovalOption: installCfg.ApprovalOption.APIType(),
			})
			if err != nil {
				return nil, err
			}
		} else {
			if appInstall.InstallConfig.ApprovalOption != installCfg.ApprovalOption.APIType() {
				// Update the install config if the approval option has changed.
				_, err := s.api.UpdateInstallConfig(ctx, appInstall.ID, appInstall.InstallConfig.ID, &models.ServiceUpdateInstallConfigRequest{
					ApprovalOption: installCfg.ApprovalOption.APIType(),
				})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	if appInstall.Metadata["managed_by"] != ManagedByNuonCLIConfig {
		_, err = s.api.UpdateInstall(ctx, appInstall.ID, &models.ServiceUpdateInstallRequest{
			Name: appInstall.Name,
			Metadata: &models.HelpersInstallMetadata{
				ManagedBy: ManagedByNuonCLIConfig,
			},
		})
	}

	// Use the current inputs as defaults, for missing values in the current inputs.
	installCfg.InputGroups = append([]config.InputGroup{currInputs.Values}, installCfg.InputGroups...)

	installCfgInputs := installCfg.FlattenedInputs()

	hasInputChanged := false
	if len(installCfg.InputGroups) != len(currInputs.Values) {
		hasInputChanged = true
	} else {
		// length is same, go through each input to see if any have changed.
		for k, v := range installCfgInputs {
			if currInputs.Values[k] != v {
				hasInputChanged = true
				break
			}
		}
	}

	// If inputs have divereged, update the install inputs.
	if hasInputChanged {
		_, workflowID, err := s.api.UpdateInstallInputs(ctx, appInstall.ID, &models.ServiceUpdateInstallInputsRequest{
			Inputs: installCfgInputs,
		})
		if err != nil {
			return nil, fmt.Errorf("error updating inputs for install %s: %w", appInstall.Name, err)
		}

		s.printWorkflowStatus(ctx, workflowID, appInstall.ID, wait)
	}

	ui.PrintSuccess(fmt.Sprintf("install %s updated successfully", appInstall.Name))
	return appInstall, nil
}

func (s *appInstallSyncer) printWorkflowStatus(ctx context.Context, workflowID string, installID string, wait bool) {
	workflow, err := s.api.GetWorkflow(ctx, workflowID)
	if err != nil {
		return
	}

	view := ui.NewGetView()
	view.Render(formatWorkflows([]*models.AppWorkflow{workflow}))

	ui.PrintWarning("Some workflow steps might need manual approval from the UI")

	if wait == false {
		return
	}

	spinner := ui.NewSpinnerView(false)
	spinner.Start("waiting for the workflow to complete")

	for !workflow.Finished {
		spinner.Update(fmt.Sprintf("waiting for the workflow to complete (status: %s)", workflow.Status.Status))

		time.Sleep(defaultPollDuration)
		currentWorkflow, err := s.api.GetWorkflow(ctx, workflowID)
		if err == nil {
			workflow = currentWorkflow
		} else {
			ui.PrintError(fmt.Errorf("failed fetching workflow status: %w", err))
		}
	}

	switch workflow.Status.Status {
	case models.AppStatusSuccess:
		spinner.Success("workflow successfully completed")
	case models.AppStatusError:
		spinner.Fail(fmt.Errorf("workflow failed"))
		cfg, err := s.api.GetCLIConfig(ctx)
		if err == nil {
			url := fmt.Sprintf(
				"%s/%s/installs/%s/workflows/%s", cfg.DashboardURL, s.orgID, installID, workflowID)
			browser.OpenURL(url)
		}
	default:
		spinner.Fail(fmt.Errorf("unknown workflow status: %s", workflow.Status.Status))
	}
}
