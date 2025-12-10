package cloudformation

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
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
