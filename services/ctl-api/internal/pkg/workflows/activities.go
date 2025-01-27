package workflows

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/activities"
	jobactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job/activities"
)

type Params struct {
	fx.In

	Activities    *activities.Activities
	JobActivities *jobactivities.Activities
}

type Activities struct {
	JobActivities *jobactivities.Activities
	Activities    *activities.Activities
}

func (a *Activities) AllActivities() []any {
	return []any{
		a.JobActivities,
		a.Activities,
	}
}

func NewActivities(params Params) *Activities {
	return &Activities{
		Activities:    params.Activities,
		JobActivities: params.JobActivities,
	}
}
