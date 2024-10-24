package log

import (
	"github.com/hashicorp/go-hclog"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/pkg/zaphclog"
)

type Params struct {
	fx.In

	L *zap.Logger `name:"system"`
}

// NOTE(jm): this will be deprecated once rolled out to each job
func SystemHclog(params Params) hclog.Logger {
	return zaphclog.Wrap(params.L)
}

func NewHClog(l *zap.Logger) hclog.Logger {
	return zaphclog.Wrap(l)
}
