package workflows

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/activities"
	jobactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/job/activities"
	signalsactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/signals/activities"
	statusactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/status/activities"
	flowactivities "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/workflows/workflow/activities"
)

type Params struct {
	fx.In

	Activities        *activities.Activities
	JobActivities     *jobactivities.Activities
	FlowActivities    *flowactivities.Activities
	SignalsActivities *signalsactivities.Activities
	StatusActivities  *statusactivities.Activities
}

type Activities struct {
	JobActivities     *jobactivities.Activities
	FlowActivities    *flowactivities.Activities
	SignalsActivities *signalsactivities.Activities
	StatusActivities  *statusactivities.Activities
	Activities        *activities.Activities
}

func (a *Activities) AllActivities() []any {
	return []any{
		a.JobActivities,
		a.FlowActivities,
		a.Activities,
		a.SignalsActivities,
		a.StatusActivities,
	}
}

func NewActivities(params Params) *Activities {
	return &Activities{
		Activities:        params.Activities,
		JobActivities:     params.JobActivities,
		FlowActivities:    params.FlowActivities,
		SignalsActivities: params.SignalsActivities,
		StatusActivities:  params.StatusActivities,
	}
}
