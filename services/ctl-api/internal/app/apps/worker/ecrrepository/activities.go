package ecrrepository

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal"
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
