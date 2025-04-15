package workflows

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/activities"
	jobactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job/activities"
	signalsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/signals/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
)

type Params struct {
	fx.In

	Activities        *activities.Activities
	JobActivities     *jobactivities.Activities
	SignalsActivities *signalsactivities.Activities
	StatusActivities  *statusactivities.Activities
}

type Activities struct {
	JobActivities     *jobactivities.Activities
	SignalsActivities *signalsactivities.Activities
	StatusActivities  *statusactivities.Activities
	Activities        *activities.Activities
}

func (a *Activities) AllActivities() []any {
	return []any{
		a.JobActivities,
		a.Activities,
		a.SignalsActivities,
		a.StatusActivities,
	}
}

func NewActivities(params Params) *Activities {
	return &Activities{
		Activities:        params.Activities,
		JobActivities:     params.JobActivities,
		SignalsActivities: params.SignalsActivities,
		StatusActivities:  params.StatusActivities,
	}
}
