package log

import (
	"github.com/hashicorp/go-hclog"
	wrapper "github.com/zaffka/zap-to-hclog"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	L *zap.Logger `name:"system"`
}

// NOTE(jm): this will be deprecated once rolled out to each job
func SystemHclog(params Params) hclog.Logger {
	return wrapper.Wrap(params.L)
}

func NewHClog(l *zap.Logger) hclog.Logger {
	return wrapper.Wrap(l)
}
