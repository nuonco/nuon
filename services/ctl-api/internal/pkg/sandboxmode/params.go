package sandboxmode

import "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"

func (s *Signal) WithParams(params *signal.Params) {
	if su, ok := s.Signal.(signal.SignalWithParams); ok {
		su.WithParams(params)
	}
}
