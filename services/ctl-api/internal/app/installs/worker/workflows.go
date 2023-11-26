package worker

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/adapters/protos"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
)

type Workflows struct {
	cfg    *internal.Config
	v      *validator.Validate
	acts   activities.Activities
	protos *protos.Adapter
}

func NewWorkflows(v *validator.Validate, cfg *internal.Config, prt *protos.Adapter) *Workflows {
	return &Workflows{
		cfg:    cfg,
		v:      v,
		protos: prt,
		//  NOTE: this field is only used to be able to fetch activity methods
		acts: activities.Activities{},
	}
}
