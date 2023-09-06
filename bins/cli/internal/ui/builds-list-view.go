package ui

import "github.com/powertoolsdev/mono/pkg/api/client/models"

type BuildsListView struct {
	ListView
}

func NewBuildsListView() *BuildsListView {
	return &BuildsListView{
		*NewListView([]string{
			"status",
			"id",
			"component id",
			"git ref",
		}),
	}
}

func (v *BuildsListView) Render(builds []*models.AppComponentBuild) {
	data := [][]string{}
	for _, build := range builds {
		data = append(data, []string{
			build.Status,
			build.ID,
			build.ComponentConfigConnectionID,
			build.GitRef,
		})
	}
	v.ListView.Render(data)
}
