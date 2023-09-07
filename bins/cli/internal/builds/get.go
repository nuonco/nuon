package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, compID, buildID string) {
	view := ui.NewGetView()

	build, err := s.api.GetComponentBuild(ctx, compID, buildID)
	if err != nil {
		view.Error(err)
		return
	}

	view.Render([][]string{
		[]string{"id", build.ID},
		[]string{"status", build.Status},
		[]string{"created at", build.CreatedAt},
		[]string{"updated at", build.UpdatedAt},
		[]string{"created by", build.CreatedByID},
		[]string{"component id", build.ComponentConfigConnectionID},

		[]string{"vcs connection id", build.VcsConnectionCommit.ID},
		[]string{"commit sha", build.VcsConnectionCommit.Sha},
		[]string{"commit author email", build.VcsConnectionCommit.AuthorEmail},
		[]string{"commit author name", build.VcsConnectionCommit.AuthorName},
		[]string{"commit created at", build.VcsConnectionCommit.CreatedAt},
		[]string{"commit updated at", build.VcsConnectionCommit.UpdatedAt},
		[]string{"commit created by", build.VcsConnectionCommit.CreatedByID},
		[]string{"commit message", build.VcsConnectionCommit.Message},
	})
}
