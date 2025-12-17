package cloudformation

import (
	"go.uber.org/fx"

	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type Templates struct {
	cfg *internal.Config
}

type Params struct {
	fx.In

	Cfg *internal.Config
}

func NewTemplates(params Params) *Templates {
	return &Templates{
		cfg: params.Cfg,
	}
}
