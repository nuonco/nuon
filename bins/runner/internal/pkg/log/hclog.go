package log

import (
	"github.com/hashicorp/go-hclog"
	wrapper "github.com/zaffka/zap-to-hclog"
	"go.uber.org/zap"
)

// NOTE(jm): this is primarily a bridge until we overhaul our logging in a more consistent way throughout the mono repo.
func NewHclog(log *zap.Logger) hclog.Logger {
	return wrapper.Wrap(log)
}
