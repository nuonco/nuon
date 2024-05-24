package builds

import (
	"context"
	"fmt"

	"github.com/mitchellh/go-wordwrap"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, appID, compID, buildID string, asJSON bool) {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
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

	vcsConnectionID := ""
	commitSha := ""
	commitAuthorEmail := ""
	commitAuthorName := ""
	commitCreatedAt := ""
	commitUpdatedAt := ""
	commitCreatedBy := ""
	commitMessage := ""
	if build.VcsConnectionCommit != nil {
		vcsConnectionID = build.VcsConnectionCommit.ID
		commitSha = build.VcsConnectionCommit.Sha
		commitAuthorEmail = build.VcsConnectionCommit.AuthorEmail
		commitAuthorName = build.VcsConnectionCommit.AuthorName
		commitCreatedAt = build.VcsConnectionCommit.CreatedAt
		commitUpdatedAt = build.VcsConnectionCommit.UpdatedAt
		commitCreatedBy = build.VcsConnectionCommit.CreatedByID
		commitMessage = build.VcsConnectionCommit.Message
	}

	buildRes := [][]string{
		{"id", build.ID},
		{"status", build.Status},
		{"created at", build.CreatedAt},
		{"updated at", build.UpdatedAt},
		{"created by", build.CreatedByID},
		{"component id", build.ComponentID},
		{"component config version", fmt.Sprintf("%d", build.ComponentConfigVersion)},

		{"vcs connection id", vcsConnectionID},
		{"commit sha", commitSha},
		{"commit author email", commitAuthorEmail},
		{"commit author name", commitAuthorName},
		{"commit created at", commitCreatedAt},
		{"commit updated at", commitUpdatedAt},
		{"commit created by", commitCreatedBy},
		{"commit message", commitMessage},

		{"description", wordwrap.WrapString(build.StatusDescription, 75)},
	}

	view.Render(buildRes)
}
