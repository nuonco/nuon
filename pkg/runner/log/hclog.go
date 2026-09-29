package log

import (
	"github.com/hashicorp/go-hclog"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/zaphclog"
)

type Params struct {
	fx.In

	L *zap.Logger `name:"system"`
}

func SystemHclog(params Params) hclog.Logger {
	return zaphclog.Wrap(params.L)
}

func NewHClog(l *zap.Logger) hclog.Logger {
	return zaphclog.Wrap(l)
}
