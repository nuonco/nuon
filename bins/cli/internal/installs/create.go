package installs

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon-go/models"
	"github.com/pkg/browser"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
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
