package temporal

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type Params struct {
	fx.In

	EVClient eventloop.Client
	L        *zap.Logger
}

type Client interface {
	Send(ctx workflow.Context, id string, signal eventloop.Signal)
	Cancel(ctx workflow.Context, namespace, id string)
}

var _ Client = (*evClient)(nil)

type evClient struct {
	evClient eventloop.Client
	l        *zap.Logger
}

func New(params Params) Client {
	return &evClient{
		evClient: params.EVClient,
		l:        params.L,
	}
}
