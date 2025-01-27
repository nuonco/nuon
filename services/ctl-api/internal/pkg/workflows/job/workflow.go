package job

import (
	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"

	teventloop "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/temporal"
)

// this is a workflow that is used to execute a job. It is designed to be reusable outside the context of this
// namespace, and for all jobs. Thus, it has it's own activities, and other components to allow it to work more
// effectively.
type Workflows struct {
	evClient teventloop.Client
}

type Params struct {
	fx.In

	V        *validator.Validate
	EVClient teventloop.Client
}

func New(params Params) (*Workflows, error) {
	return &Workflows{
		evClient: params.EVClient,
	}, nil
}
