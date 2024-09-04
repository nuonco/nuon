package temporal

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	EVClient eventloop.Client
}

type Client interface {
	Send(ctx workflow.Context, id string, signal eventloop.Signal)
}

var _ Client = (*evClient)(nil)

type evClient struct {
	evClient eventloop.Client
}

func New(params Params) Client {
	return &evClient{
		evClient: params.EVClient,
	}
}
