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
	releasePackageStatusActive = "active"
	releasePackageStatusError  = "error"
	bundlePollInterval         = 5 * time.Second
	bundleWaitTimeout          = 60 * time.Minute
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
	release, err := s.api.CreateAppRelease(ctx, appID, appConfigID)
	if err != nil {
		return ui.NewGetView().Error(err)
	}
	if release == nil {
		return ui.PrintError(fmt.Errorf("create app release returned an empty response"))
	}
	pkg, err := s.api.CreateReleasePackage(ctx, appID, release.ID, platform)
	if err != nil {
		return ui.NewGetView().Error(err)
	}
	if pkg == nil {
		return ui.PrintError(fmt.Errorf("create release package returned an empty response"))
	}
	if pkg.Status != releasePackageStatusActive && !opts.NoWait {
		pkg, err = s.waitForReleasePackage(ctx, pkg, opts.PrintJSON)
		if err != nil {
			return ui.PrintError(err)
		}
	}
	if opts.PrintJSON {
		ui.PrintJSON(pkg)
		return nil
	}
	renderBundle(release, pkg)
	return nil
}

func (s *Service) waitForReleasePackage(ctx context.Context, pkg *models.AppReleasePackage, quiet bool) (*models.AppReleasePackage, error) {
	waitCtx, cancel := context.WithTimeout(ctx, bundleWaitTimeout)
	defer cancel()

	lastStatus := ""
	for {
		if pkg.Status != lastStatus && !quiet {
			ui.PrintLn(fmt.Sprintf("bundle %s status: %s", pkg.ID, pkg.Status))
			lastStatus = pkg.Status
		}
		switch pkg.Status {
		case releasePackageStatusActive:
			return pkg, nil
		case releasePackageStatusError:
			description := pkg.StatusDescription
			if description == "" {
				description = "bundle publishing failed"
			}
			return nil, fmt.Errorf("bundle %s failed: %s", pkg.ID, description)
		}

		select {
		case <-waitCtx.Done():
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, fmt.Errorf("timed out after %s waiting for bundle %s; resume with `nuon apps bundles get %s`", bundleWaitTimeout, pkg.ID, pkg.ID)
		case <-time.After(bundlePollInterval):
		}

		var err error
		pkg, err = s.api.GetReleasePackage(waitCtx, pkg.ID)
		if err != nil {
			return nil, fmt.Errorf("check bundle %s status: %w", pkg.ID, err)
		}
	}
}

func renderBundle(release *models.AppAppRelease, pkg *models.AppReleasePackage) {
	ui.NewGetView().Render([][]string{
		{"id", pkg.ID},
		{"release id", release.ID},
		{"app config id", release.AppConfigID},
		{"target platform", pkg.TargetPlatform},
		{"size (bytes)", strconv.FormatInt(pkg.ArchiveSize, 10)},
		{"release digest", release.SemanticDigest},
		{"package digest", pkg.PackageDigest},
		{"manifest digest", pkg.ManifestDigest},
	})
}

func (s *Service) ListBundles(ctx context.Context, appID string, offset, limit int, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	releases, hasMore, err := s.api.ListAppReleases(ctx, appID, &models.GetPaginatedQuery{Offset: offset, Limit: limit})
	if err != nil {
		return ui.NewListView().Error(err)
	}
	packages := make([]*models.AppReleasePackage, 0)
	for _, release := range releases {
		packages = append(packages, release.Packages...)
	}
	if asJSON {
		ui.PrintJSON(packages)
		return nil
	}

	data := [][]string{{"ID", "RELEASE ID", "TARGET PLATFORM", "STATUS", "SIZE (BYTES)", "CREATED AT"}}
	for _, pkg := range packages {
		data = append(data, []string{pkg.ID, pkg.ReleaseID, pkg.TargetPlatform, pkg.Status, strconv.FormatInt(pkg.ArchiveSize, 10), pkg.CreatedAt})
	}
	ui.NewListView().RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) GetBundle(ctx context.Context, appID, bundleID string, asJSON bool) error {
	appID, err := s.bundleAppID(ctx, appID)
	if err != nil {
		return ui.PrintError(err)
	}
	pkg, err := s.api.GetReleasePackage(ctx, bundleID)
	if err != nil {
		return ui.NewGetView().Error(err)
	}
	release, err := s.api.GetAppRelease(ctx, appID, pkg.ReleaseID)
	if err != nil {
		return ui.NewGetView().Error(err)
	}
	if asJSON {
		ui.PrintJSON(pkg)
		return nil
	}

	ui.NewGetView().Render([][]string{
		{"id", pkg.ID},
		{"app id", release.AppID},
		{"release id", pkg.ReleaseID},
		{"app config id", release.AppConfigID},
		{"format", pkg.Format},
		{"target platform", pkg.TargetPlatform},
		{"status", pkg.Status},
		{"schema version", strconv.FormatInt(pkg.SchemaVersion, 10)},
		{"size (bytes)", strconv.FormatInt(pkg.ArchiveSize, 10)},
		{"archive checksum", pkg.ArchiveChecksum},
		{"release digest", release.SemanticDigest},
		{"package digest", pkg.PackageDigest},
		{"manifest digest", pkg.ManifestDigest},
		{"oci root digest", pkg.OciRootDigest},
		{"members", strconv.Itoa(len(pkg.Members))},
		{"created at", pkg.CreatedAt},
	})
	return nil
}
