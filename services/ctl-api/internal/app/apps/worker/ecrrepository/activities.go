package ecrrepository

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/fx"
)

type Activities struct {
	cfg *internal.Config
}

type ActivitiesParams struct {
	fx.In

	Cfg *internal.Config
}

func NewActivities(params *ActivitiesParams) *Activities {
	return &Activities{
		cfg: params.Cfg,
	}
}
