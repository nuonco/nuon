package releases

import (
	"context"
	"strconv"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, releaseID string, asJSON bool) {
	view := ui.NewGetView()

	release, err := s.api.GetRelease(ctx, releaseID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(release)
		return
	}

	view.Render([][]string{
		{"id", release.ID},
		{"status", release.Status},
		{"created at", release.CreatedAt},
		{"updated at", release.UpdatedAt},
		{"created by", release.CreatedByID},
		{"build id", release.BuildID},
		{"total steps", strconv.Itoa(int(release.TotalReleaseSteps))},
	})
}
