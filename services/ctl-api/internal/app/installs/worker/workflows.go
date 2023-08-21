package worker

import (
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"go.uber.org/zap"
)

type Workflows struct {
	cfg  *internal.Config
	v    *validator.Validate
	acts activities.Activities
	l    *zap.Logger
}

func NewWorkflows(v *validator.Validate, cfg *internal.Config, l *zap.Logger) *Workflows {
	return &Workflows{
		cfg: cfg,
		v:   v,
		l:   l,
		//  NOTE: this field is only used to be able to fetch activity methods
		acts: activities.Activities{},
	}
}
