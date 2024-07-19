package sync

import (
	"context"

	"github.com/nuonco/nuon-go"
	"github.com/nuonco/nuon-go/models"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (s *sync) lookupAppIDs(ctx context.Context, resource string) ([]string, error) {
	appIDs := make([]string, 0)

	for _, appID := range s.cfg.Installer.AppIDs {
		app, err := s.apiClient.GetApp(ctx, appID)
		if err == nil {
			appIDs = append(appIDs, app.ID)
			continue
		}

		if !nuon.IsNotFound(err) {
			return nil, SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}

		return nil, SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}

	return appIDs, nil
}

func (s *sync) syncAppInstaller(ctx context.Context, resource string) error {
	if s.cfg.Installer == nil {
		return nil
	}

	appIDs, err := s.lookupAppIDs(ctx, resource)

	if s.prevState.InstallerID != "" {
		req := s.updateInstallerRequest(appIDs)

		_, err = s.apiClient.UpdateInstaller(ctx, s.prevState.InstallerID, req)
		if err != nil {
			return SyncAPIErr{
				Resource: resource,
				Err:      err,
			}
		}

		return nil
	}

	req := s.createInstallerRequest(appIDs)
	installer, err := s.apiClient.CreateInstaller(ctx, req)
	if err != nil {
		return SyncAPIErr{
			Resource: resource,
			Err:      err,
		}
	}
	s.state.InstallerID = installer.ID

	return nil
}

func (s *sync) createInstallerRequest(appIDs []string) *models.ServiceCreateInstallerRequest {
	return &models.ServiceCreateInstallerRequest{
		AppIds: appIDs,
		Name:   generics.ToPtr(s.cfg.Installer.Name),
		Metadata: &models.ServiceCreateInstallerRequestMetadata{
			CommunityURL:        generics.ToPtr(s.cfg.Installer.CommunityURL),
			CopyrightMarkdown:   s.cfg.Installer.FooterMarkdown,
			DemoURL:             s.cfg.Installer.DemoURL,
			Description:         generics.ToPtr(s.cfg.Installer.Description),
			DocumentationURL:    generics.ToPtr(s.cfg.Installer.DocumentationURL),
			FaviconURL:          generics.ToPtr(s.cfg.Installer.FaviconURL),
			FooterMarkdown:      s.cfg.Installer.FooterMarkdown,
			HomepageURL:         generics.ToPtr(s.cfg.Installer.HomepageURL),
			GithubURL:           generics.ToPtr(s.cfg.Installer.GithubURL),
			LogoURL:             generics.ToPtr(s.cfg.Installer.LogoURL),
			OgImageURL:          s.cfg.Installer.OgImageURL,
			PostInstallMarkdown: s.cfg.Installer.PostInstallMarkdown,
		},
	}
}

func (s *sync) updateInstallerRequest(appIDs []string) *models.ServiceUpdateInstallerRequest {
	return &models.ServiceUpdateInstallerRequest{
		AppIds: appIDs,
		Name:   &s.cfg.Installer.Name,
		Metadata: &models.ServiceUpdateInstallerRequestMetadata{
			CommunityURL:        generics.ToPtr(s.cfg.Installer.CommunityURL),
			CopyrightMarkdown:   s.cfg.Installer.FooterMarkdown,
			DemoURL:             s.cfg.Installer.DemoURL,
			Description:         generics.ToPtr(s.cfg.Installer.Description),
			DocumentationURL:    generics.ToPtr(s.cfg.Installer.DocumentationURL),
			FaviconURL:          generics.ToPtr(s.cfg.Installer.FaviconURL),
			FooterMarkdown:      s.cfg.Installer.FooterMarkdown,
			HomepageURL:         generics.ToPtr(s.cfg.Installer.HomepageURL),
			GithubURL:           generics.ToPtr(s.cfg.Installer.GithubURL),
			LogoURL:             generics.ToPtr(s.cfg.Installer.LogoURL),
			OgImageURL:          s.cfg.Installer.OgImageURL,
			PostInstallMarkdown: s.cfg.Installer.PostInstallMarkdown,
		},
	}
}
