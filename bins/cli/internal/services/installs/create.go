package installs

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/pkg/browser"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/services/labels"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	appselector "github.com/nuonco/nuon/bins/cli/internal/ui/v3/app/selector"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/install/creator"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
)

const (
	statusError       = "error"
	statusActive      = "active"
	statusAccessError = "access_error"
)

func (s *Service) Create(ctx context.Context, appID, name, region string, inputs, labelArgs []string, asJSON, noSelect bool) error {
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

	// we collect these and pass them down so we can pre-fill specific fields
	inputsMap := make(map[string]string)
	for _, kv := range inputs {
		kvT := strings.Split(kv, "=")
		inputsMap[kvT[0]] = kvT[1]
	}

	labelsMap, removeKeys, err := labels.ParseArgs(labelArgs)
	if err != nil {
		return ui.PrintError(err)
	}
	if len(removeKeys) > 0 {
		return ui.PrintError(fmt.Errorf("label removal (key-) is not allowed at install creation; use `nuon installs label` after the install exists"))
	}

	req, err := s.buildCreateInstallRequest(ctx, appID, name, region, inputsMap, labelsMap)
	if err != nil {
		return ui.PrintError(err)
	}

	if asJSON {
		install, err := s.api.CreateInstall(ctx, appID, req)
		if err != nil {
			return ui.PrintJSONError(err)
		}
		ui.PrintJSON(install)
		return nil
	}

	if s.cfg.Preview {
		installID, _ := creator.InstallCreatorApp(
			ctx,
			s.cfg,
			s.api,
			appID,
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
		wf := workflows[0]
		workflow.WorkflowApp(ctx, s.cfg, s.api, installID, wf.ID, false)
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

func (s *Service) buildCreateInstallRequest(ctx context.Context, appID, name, region string, inputs, labelsMap map[string]string) (*models.ServiceCreateInstallRequest, error) {
	req := &models.ServiceCreateInstallRequest{
		Name:   &name,
		Inputs: s.inputsWithDefaults(ctx, appID, inputs),
		Labels: labelsMap,
	}

	runnerCfg, err := s.api.GetAppRunnerLatestConfig(ctx, appID)
	if err != nil || runnerCfg == nil {
		req.AwsAccount = &models.ServiceCreateInstallRequestAwsAccount{Region: region}
		return req, nil
	}

	switch runnerCfg.CloudPlatform {
	case models.AppCloudPlatformGcp:
		req.GcpAccount = &models.ServiceCreateInstallRequestGcpAccount{}
	case models.AppCloudPlatformAzure:
		req.AzureAccount = &models.ServiceCreateInstallRequestAzureAccount{}
	default:
		if region == "" {
			return nil, fmt.Errorf("--region is required for AWS installs")
		}
		req.AwsAccount = &models.ServiceCreateInstallRequestAwsAccount{Region: region}
	}

	return req, nil
}

// inputsWithDefaults merges app input defaults with any explicitly provided values.
// Explicit values win; defaults fill in anything not provided.
func (s *Service) inputsWithDefaults(ctx context.Context, appID string, provided map[string]string) map[string]string {
	inputCfg, err := s.api.GetAppInputLatestConfig(ctx, appID)
	if err != nil || inputCfg == nil {
		return provided
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
	return merged
}
