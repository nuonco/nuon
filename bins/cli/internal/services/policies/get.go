package policies

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, appID, configID string, asJSON bool) error {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	var cfg interface{}
	if configID == "" {
		cfg, err = s.api.GetLatestAppPoliciesConfig(ctx, appID)
	} else {
		cfg, err = s.api.GetAppPoliciesConfig(ctx, appID, configID)
	}

	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch policies config"))
	}

	if asJSON {
		ui.PrintJSON(cfg)
		return nil
	}

	// Type assert to get the actual config
	policiesConfig, ok := cfg.(*struct {
		ID        string
		CreatedAt string
		Policies  interface{}
	})
	if !ok {
		// Fallback to JSON output for complex types
		ui.PrintJSON(cfg)
		return nil
	}

	data := [][]string{
		{"id", policiesConfig.ID},
		{"created at", policiesConfig.CreatedAt},
		{"policies", fmt.Sprintf("%v", policiesConfig.Policies)},
	}

	view.Render(data)
	return nil
}
