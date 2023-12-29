package releases

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mitchellh/go-wordwrap"
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

	// render out individual installs
	data := make([][]string, 0)
	data = append(data, []string{
		"install id",
		"deploy id",
		"step",
		"status",
		"description",
	})
	for idx, step := range release.ReleaseSteps {
		for _, installDeploy := range step.InstallDeploys {
			data = append(data, []string{
				installDeploy.InstallID,
				installDeploy.ID,
				fmt.Sprintf("%d", idx),
				installDeploy.Status,
				wordwrap.WrapString(installDeploy.StatusDescription, 75),
			})
		}
	}

	view.Render(data)
}
