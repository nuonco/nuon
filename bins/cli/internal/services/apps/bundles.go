package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type BundleListOptions struct {
	ConfigID string
	Status   string
	Offset   int
	Limit    int
}

func (s *Service) bundleAppID(ctx context.Context, appID string) (string, error) {
	if appID == "" {
		appID = s.getAppID()
	}
	return lookup.AppID(ctx, s.api, appID)
}

func (s *Service) ListBundles(ctx context.Context, appID string, opts BundleListOptions, asJSON bool) error {
	if opts.Offset < 0 || opts.Limit < 1 || opts.Limit > 100 {
		return ui.PrintError(&ui.CLIUserError{Msg: "offset must be non-negative and limit must be between 1 and 100"})
	}
	switch opts.Status {
	case "", "queued", "publishing", "active", "error":
	default:
		return ui.PrintError(&ui.CLIUserError{Msg: "status must be queued, publishing, active, or error"})
	}
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	bundles, hasMore, err := s.api.GetAppBundles(ctx, appID, &nuon.GetAppBundlesQuery{
		Pagination:  models.GetPaginatedQuery{Offset: opts.Offset, Limit: opts.Limit},
		AppConfigID: opts.ConfigID, Status: opts.Status,
	})
	if err != nil {
		return ui.PrintError(err)
	}
	if asJSON {
		ui.PrintJSON(bundles)
		return nil
	}
	rows := [][]string{{"ID", "CONFIG ID", "PLATFORM", "STATUS", "SIZE (BYTES)", "CREATED AT"}}
	for _, bundle := range bundles {
		rows = append(rows, []string{bundle.ID, bundle.AppConfigID, bundle.TargetPlatform, bundle.Status, fmt.Sprint(bundle.Size), bundle.CreatedAt})
	}
	ui.NewListView().RenderPaging(rows, opts.Offset, opts.Limit, hasMore)
	return nil
}

func (s *Service) GetBundle(ctx context.Context, appID, bundleID string, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	bundle, err := s.api.GetAppBundle(ctx, appID, bundleID)
	if err != nil {
		return ui.PrintError(err)
	}
	return renderBundle(bundle, asJSON)
}

func (s *Service) CreateBundle(ctx context.Context, appID, configID, platform string, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	if configID == "" {
		cfg, err := s.api.GetAppLatestConfig(ctx, appID)
		if err != nil {
			return ui.PrintError(err)
		}
		if cfg.Status != "active" {
			return ui.PrintError(&ui.CLIUserError{Msg: fmt.Sprintf("latest config %s has status %s; sync it successfully or select an active config with --config-id", cfg.ID, cfg.Status)})
		}
		configID = cfg.ID
	}
	bundle, err := s.api.CreateAppBundle(ctx, appID, &models.ServiceCreateBundleRequest{AppConfigID: &configID, TargetPlatform: platform})
	if err != nil {
		return ui.PrintError(err)
	}
	if err := renderBundle(bundle, asJSON); err != nil {
		return err
	}
	if !asJSON {
		ui.Printf("\nWait:     nuon apps bundles wait --app-id %s --bundle-id %s\n", appID, bundle.ID)
		ui.Printf("Download: nuon apps bundles download --app-id %s --bundle-id %s\n", appID, bundle.ID)
	}
	return nil
}

func (s *Service) WaitBundle(ctx context.Context, appID, bundleID string, timeout time.Duration, asJSON bool) error {
	if timeout <= 0 {
		return ui.PrintError(&ui.CLIUserError{Msg: "timeout must be positive"})
	}
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	if !asJSON {
		ui.Printf("Waiting for bundle %s. Ctrl+C stops waiting, not publication.\n", bundleID)
	}
	spinner := bubbles.NewSpinnerView(asJSON, s.cfg.Interactive)
	spinner.Start("Waiting for publication")
	progress := func(bundle *models.ServiceBundleResponse) {
		spinner.Update(fmt.Sprintf("Bundle %s: %s — %s", bundle.ID, bundle.Status, bundle.StatusDescription))
	}
	bundle, err := s.waitBundle(ctx, appID, bundleID, timeout, progress)
	if err != nil {
		if !asJSON {
			spinner.Fail(err)
		}
		return ui.PrintError(err)
	}
	if !asJSON {
		spinner.Success("Bundle published and verified")
	}
	return renderBundle(bundle, asJSON)
}

func (s *Service) waitBundle(ctx context.Context, appID, bundleID string, timeout time.Duration, progress func(*models.ServiceBundleResponse)) (*models.ServiceBundleResponse, error) {
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		bundle, err := s.api.GetAppBundle(pollCtx, appID, bundleID)
		if pollCtx.Err() != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, &ui.CLIUserError{Msg: fmt.Sprintf("bundle %s timed out after %s; publication continues. Resume: nuon apps bundles wait --app-id %s --bundle-id %s", bundleID, timeout, appID, bundleID)}
		}
		if err != nil {
			return nil, err
		}
		progress(bundle)
		switch bundle.Status {
		case "active":
			return bundle, nil
		case "error":
			return nil, &ui.CLIUserError{Msg: fmt.Sprintf("bundle %s publication failed: %s", bundleID, bundle.StatusDescription)}
		case "queued", "publishing":
		default:
			return nil, &ui.CLIUserError{Msg: fmt.Sprintf("bundle %s has unknown status %q", bundleID, bundle.Status)}
		}
		select {
		case <-pollCtx.Done():
		case <-ticker.C:
		}
	}
}

func renderBundle(bundle *models.ServiceBundleResponse, asJSON bool) error {
	if asJSON {
		ui.PrintJSON(bundle)
		return nil
	}
	ui.NewGetView().Render([][]string{
		{"id", bundle.ID}, {"app_id", bundle.AppID}, {"config_id", bundle.AppConfigID},
		{"platform", bundle.TargetPlatform}, {"status", bundle.Status}, {"description", bundle.StatusDescription},
		{"size_bytes", fmt.Sprint(bundle.Size)}, {"created_at", bundle.CreatedAt},
		{"verified_at", bundle.VerifiedAt}, {"checksum", bundle.TransportChecksum},
		{"manifest_digest", bundle.ManifestDigest},
	})
	return nil
}
