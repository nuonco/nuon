package service

import (
	"context"
	"fmt"

	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (s *service) hydrateBuildRunnerJobs(ctx context.Context, builds ...*app.ComponentBuild) error {
	if len(builds) == 0 {
		return nil
	}

	buildIDs := make([]any, 0, len(builds))
	buildsByID := make(map[string]*app.ComponentBuild, len(builds))
	for _, build := range builds {
		buildIDs = append(buildIDs, build.ID)
		buildsByID[build.ID] = build
	}

	jobs := []app.RunnerJob{}
	if err := s.db.WithContext(ctx).
		Scopes(scopes.WithDisableViews).
		Where(&app.RunnerJob{
			OwnerType: "component_builds",
			Group:     app.RunnerJobGroupBuild,
			Operation: app.RunnerJobOperationTypeBuild,
		}).
		Where(clause.IN{Column: "owner_id", Values: buildIDs}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}).
		Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
		Find(&jobs).Error; err != nil {
		return fmt.Errorf("unable to load component build execution jobs: %w", err)
	}

	for i := range jobs {
		build := buildsByID[jobs[i].OwnerID]
		if build.BuildRunnerJobID != nil {
			continue
		}
		build.RunnerJob = jobs[i]
		build.BuildRunnerJobID = &jobs[i].ID
	}

	return nil
}
