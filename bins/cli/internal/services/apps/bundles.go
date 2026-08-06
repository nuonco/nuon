package apps

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

const (
	airgapBundleStatusActive = "active"
	airgapBundleStatusError  = "error"
	airgapBundlePollInterval = 5 * time.Second
	airgapBundleWaitTimeout  = 60 * time.Minute
)

type CreateBundleOptions struct {
	NoWait    bool
	PrintJSON bool
}

func (s *Service) bundleAppID(ctx context.Context, appID string) (string, error) {
	if appID == "" {
		appID = s.cfg.GetString("app_id")
	}
	return lookup.AppID(ctx, s.api, appID)
}

func (s *Service) CreateBundle(ctx context.Context, appID, appConfigID, platform string, opts CreateBundleOptions) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	bundle, err := s.api.CreateAirgapBundle(ctx, appID, appConfigID, platform)
	if err != nil {
		return ui.NewGetView().Error(err)
	}
	if bundle == nil {
		return ui.PrintError(fmt.Errorf("create air-gap bundle returned an empty response"))
	}
	if bundle.Status != airgapBundleStatusActive && !opts.NoWait {
		bundle, err = s.waitForAirgapBundle(ctx, appID, bundle, opts.PrintJSON)
		if err != nil {
			return ui.PrintError(err)
		}
	}
	if opts.PrintJSON {
		ui.PrintJSON(bundle)
		return nil
	}
	renderBundle(bundle)
	return nil
}

func (s *Service) waitForAirgapBundle(ctx context.Context, appID string, bundle *models.ServiceBundleResponse, quiet bool) (*models.ServiceBundleResponse, error) {
	waitCtx, cancel := context.WithTimeout(ctx, airgapBundleWaitTimeout)
	defer cancel()

	lastStatus := ""
	for {
		if bundle.Status != lastStatus && !quiet {
			ui.PrintLn(fmt.Sprintf("bundle %s status: %s", bundle.ID, bundle.Status))
			lastStatus = bundle.Status
		}
		switch bundle.Status {
		case airgapBundleStatusActive:
			return bundle, nil
		case airgapBundleStatusError:
			description := bundle.StatusDescription
			if description == "" {
				description = "bundle publishing failed"
			}
			return nil, fmt.Errorf("air-gap bundle %s failed: %s", bundle.ID, description)
		}

		select {
		case <-waitCtx.Done():
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("timed out after %s waiting for air-gap bundle %s; resume with `nuon apps bundles get %s`", airgapBundleWaitTimeout, bundle.ID, bundle.ID)
		case <-time.After(airgapBundlePollInterval):
		}

		var err error
		bundle, err = s.api.GetAirgapBundle(waitCtx, appID, bundle.ID)
		if err != nil {
			return nil, fmt.Errorf("check air-gap bundle %s status: %w", bundle.ID, err)
		}
	}
}

func renderBundle(bundle *models.ServiceBundleResponse) {
	ui.NewGetView().Render([][]string{{"id", bundle.ID}, {"app config id", bundle.AppConfigID}, {"target platform", bundle.TargetPlatform}, {"size (bytes)", strconv.FormatInt(bundle.Size, 10)}, {"manifest digest", bundle.ManifestDigest}})
}

func (s *Service) ListBundles(ctx context.Context, appID string, offset, limit int, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	bundles, hasMore, err := s.api.ListAirgapBundles(ctx, appID, &models.GetPaginatedQuery{Offset: offset, Limit: limit})
	if err != nil {
		return ui.NewListView().Error(err)
	}
	if asJSON {
		ui.PrintJSON(bundles)
		return nil
	}

	data := [][]string{{"ID", "APP CONFIG ID", "TARGET PLATFORM", "SIZE (BYTES)", "CREATED AT"}}
	for _, bundle := range bundles {
		data = append(data, []string{bundle.ID, bundle.AppConfigID, bundle.TargetPlatform, strconv.FormatInt(bundle.Size, 10), bundle.CreatedAt})
	}
	ui.NewListView().RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) GetBundle(ctx context.Context, appID, bundleID string, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	bundle, err := s.api.GetAirgapBundle(ctx, appID, bundleID)
	if err != nil {
		return ui.NewGetView().Error(err)
	}
	if asJSON {
		ui.PrintJSON(bundle)
		return nil
	}

	ui.NewGetView().Render([][]string{
		{"id", bundle.ID},
		{"app id", bundle.AppID},
		{"app config id", bundle.AppConfigID},
		{"target platform", bundle.TargetPlatform},
		{"schema version", strconv.FormatInt(bundle.SchemaVersion, 10)},
		{"size (bytes)", strconv.FormatInt(bundle.Size, 10)},
		{"transport checksum", bundle.TransportChecksum},
		{"manifest digest", bundle.ManifestDigest},
		{"oci root digest", bundle.OciRootDigest},
		{"artifacts", strconv.Itoa(len(bundle.Artifacts))},
		{"created at", bundle.CreatedAt},
	})
	return nil
}

func (s *Service) CreateBundleInstall(ctx context.Context, appID, bundleID, name string, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	install, err := s.api.CreateAirgapInstall(ctx, appID, bundleID, name)
	if err != nil {
		return ui.NewGetView().Error(err)
	}
	if asJSON {
		ui.PrintJSON(install)
		return nil
	}
	ui.NewGetView().Render([][]string{
		{"id", install.ID},
		{"name", install.Name},
		{"app id", install.AppID},
		{"app config id", install.AppConfigID},
		{"bundle id", install.AirgapBundleID},
		{"created at", install.CreatedAt},
	})
	ui.PrintLn("")
	ui.PrintLn("Add this to the customer's `nuon-bundle init` config file:")
	ui.PrintLn("")
	ui.PrintLn(fmt.Sprintf("  install_id: %s", install.ID))
	return nil
}

func (s *Service) ListBundleInstalls(ctx context.Context, appID, bundleID string, offset, limit int, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	installs, hasMore, err := s.api.ListAirgapInstalls(ctx, appID, bundleID, &models.GetPaginatedQuery{Offset: offset, Limit: limit})
	if err != nil {
		return ui.NewListView().Error(err)
	}
	if asJSON {
		ui.PrintJSON(installs)
		return nil
	}

	data := [][]string{{"ID", "NAME", "BUNDLE ID", "CREATED AT"}}
	for _, install := range installs {
		data = append(data, []string{install.ID, install.Name, install.AirgapBundleID, install.CreatedAt})
	}
	ui.NewListView().RenderPaging(data, offset, limit, hasMore)
	return nil
}
