package ui

import "github.com/powertoolsdev/mono/pkg/api/client/models"

type BuildsGetView struct {
	GetView
}

func NewBuildsGetView() *BuildsGetView {
	return &BuildsGetView{
		*NewGetView([]string{
			"id",
			"status",
			"created at",
			"updated at",
			"created by",
			"component id",

			"vcs connection id",
			"commit sha",
			"commit author email",
			"commit author name",
			"commit created at",
			"commit updated at",
			"commit created by",
			"commit message",
		}),
	}
}

func (v *BuildsGetView) Render(build *models.AppComponentBuild) {
	commit := build.VcsConnectionCommit
	data := []string{
		build.ID,
		build.StatusDescription,
		build.CreatedAt,
		build.UpdatedAt,
		build.CreatedByID,
		build.ComponentConfigConnectionID,

		commit.ID,
		commit.Sha,
		commit.AuthorEmail,
		commit.AuthorName,
		commit.CreatedAt,
		commit.UpdatedAt,
		commit.CreatedByID,
		commit.Message,

		// build.Releases,
	}
	v.GetView.Render(data)
}
