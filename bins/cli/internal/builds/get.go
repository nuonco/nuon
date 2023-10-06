package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, compID, buildID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, compID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	build, err := s.api.GetComponentBuild(ctx, compID, buildID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(build)
		return
	}

	view.Render([][]string{
		{"id", build.ID},
		{"status", build.Status},
		{"created at", build.CreatedAt},
		{"updated at", build.UpdatedAt},
		{"created by", build.CreatedByID},
		{"component id", build.ComponentConfigConnectionID},

		{"vcs connection id", build.VcsConnectionCommit.ID},
		{"commit sha", build.VcsConnectionCommit.Sha},
		{"commit author email", build.VcsConnectionCommit.AuthorEmail},
		{"commit author name", build.VcsConnectionCommit.AuthorName},
		{"commit created at", build.VcsConnectionCommit.CreatedAt},
		{"commit updated at", build.VcsConnectionCommit.UpdatedAt},
		{"commit created by", build.VcsConnectionCommit.CreatedByID},
		{"commit message", build.VcsConnectionCommit.Message},
	})
}
