package builds

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/mitchellh/go-wordwrap"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/oci/imageref"
)

func (s *Service) Get(ctx context.Context, appID, compID, buildID string, asJSON bool) error {
	compID, err := lookup.ComponentID(ctx, s.api, appID, compID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	build, err := s.api.GetComponentBuild(ctx, compID, buildID)
	if err != nil {
		return view.Error(errors.Wrap(err, "failed to fetch component build"))
	}

	if asJSON {
		ui.PrintJSON(build)
		return nil
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

	status := build.Status
	if build.NoOp {
		status = status + " (no-op)"
	}

	buildRes := [][]string{
		{"id", build.ID},
		{"status", status},
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

		// Image-source identity fields. Empty for non-image builds.
		{"source ref", build.SourceRef},
		{"source image", build.SourceImage},
		{"resolved tag", build.ResolvedTag},
		{"source digest", build.SourceDigest},
		{"source media type", build.SourceMediaType},
		{"resolved at", build.ResolvedAt},
		{"no-op", fmt.Sprintf("%t", build.NoOp)},
		{"display ref", imageref.DisplayRef(imageref.Source{
			SourceImage:  build.SourceImage,
			SourceRef:    build.SourceRef,
			ResolvedTag:  build.ResolvedTag,
			SourceDigest: build.SourceDigest,
		})},
		{"image ref", imageref.ImageRef(imageref.Source{
			SourceImage:  build.SourceImage,
			SourceDigest: build.SourceDigest,
		})},

		{"description", wordwrap.WrapString(build.StatusDescription, 75)},
	}

	view.Render(buildRes)
	return nil
}
