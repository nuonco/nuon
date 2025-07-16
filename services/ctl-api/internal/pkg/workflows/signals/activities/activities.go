package activities

import (
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	EVClient eventloop.Client
}

type Activities struct {
	evClient eventloop.Client
}

func New(params Params) *Activities {
	return &Activities{
		evClient: params.EVClient,
	}
}
