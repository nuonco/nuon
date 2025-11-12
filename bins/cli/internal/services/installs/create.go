package installs

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/browser"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/install/creator"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui/v3/workflow"
)

const (
	statusError       = "error"
	statusActive      = "active"
	statusAccessError = "access_error"
)

func (s *Service) Create(ctx context.Context, appID, name, region string, inputs []string, asJSON, noSelect bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	// we collect these and pass them down so we can pre-fill specific fields
	inputsMap := make(map[string]string)
	for _, kv := range inputs {
		kvT := strings.Split(kv, "=")
		inputsMap[kvT[0]] = kvT[1]
	}

	if asJSON {
		install, _, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
			Name: &name,
			AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
				Region: region,
			},
			Inputs: inputsMap,
		})
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
		workflow.WorkflowApp(ctx, s.cfg, s.api, installID, wf.ID)
		return nil

	}

	install, _, err := s.api.CreateInstall(ctx, appID, &models.ServiceCreateInstallRequest{
		Name: &name,
		AwsAccount: &models.ServiceCreateInstallRequestAwsAccount{
			Region: region,
		},
		Inputs: inputsMap,
	})
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
