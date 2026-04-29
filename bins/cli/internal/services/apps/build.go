package apps

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/v3/workflow"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) Build(ctx context.Context, appID, configID string) error {
	if appID == "" {
		appID = s.getAppID()
	}
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(errors.Wrap(err, "unable to resolve app"))
	}

	// If no config ID provided, use the latest
	if configID == "" {
		configs, _, err := s.api.GetAppConfigs(ctx, appID, &models.GetPaginatedQuery{Limit: 1, Offset: 0})
		if err != nil {
			return ui.PrintError(errors.Wrap(err, "unable to get app configs"))
		}
		if len(configs) == 0 {
			return ui.PrintError(fmt.Errorf("no app configs found; run 'nuon apps sync' first"))
		}
		configID = configs[0].ID
	}

	ui.PrintLn(fmt.Sprintf("building app config %s", configID))

	wf, err := s.api.BuildAppConfig(ctx, appID, configID)
	if err != nil {
		return ui.PrintError(errors.Wrap(err, "unable to start build"))
	}

	ui.PrintLn(fmt.Sprintf("workflow %s created", wf.ID))

	// Show the workflow TUI (passing empty installID since this is app-level)
	workflow.WorkflowApp(ctx, s.cfg, s.api, "", wf.ID, false)
	return nil
}
