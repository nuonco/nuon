package releases

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) List(ctx context.Context, appID, compID string) error {
	releases := []*models.AppComponentRelease{}
	err := error(nil)
	if appID != "" {
		releases, err = s.api.GetAppReleases(ctx, appID)
	} else if compID != "" {
		releases, err = s.api.GetComponentReleases(ctx, compID)
	}
	if err != nil {
		return err
	}

	if len(releases) == 0 {
		ui.Line(ctx, "No components found")
	} else {
		for _, release := range releases {
			ui.Line(ctx, "%s - %s", release.ID, release.Status)
		}
	}

	return nil
}
