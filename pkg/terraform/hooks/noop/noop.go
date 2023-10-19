package noop

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/powertoolsdev/mono/pkg/terraform/hooks"
)

type noop struct{}

var _ hooks.Hooks = (*noop)(nil)

func New() *noop {
	return &noop{}
}

func (n *noop) Init(context.Context, string) error { return nil }

func (n *noop) PreApply(context.Context, hclog.Logger) error   { return nil }
func (n *noop) PostApply(context.Context, hclog.Logger) error  { return nil }
func (n *noop) ErrorApply(context.Context, hclog.Logger) error { return nil }

func (n *noop) PreDestroy(context.Context, hclog.Logger) error   { return nil }
func (n *noop) PostDestroy(context.Context, hclog.Logger) error  { return nil }
func (n *noop) ErrorDestroy(context.Context, hclog.Logger) error { return nil }
