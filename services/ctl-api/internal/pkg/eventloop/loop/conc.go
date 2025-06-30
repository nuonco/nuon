package loop

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type concgroup[SignalType eventloop.Signal] struct {
	queue []SignalType
}

type grp[SignalType eventloop.Signal] struct {
	// groups is all the separate signal groups created for the desired distinct groups
	groups map[string]*concgroup[SignalType]
}

func (l *Loop[SignalType, ReqSig]) addGroup(name string) *concgroup[SignalType] {
	newcg := &concgroup[SignalType]{
		queue: make([]SignalType, 0, maxSignals),
	}
	l.conc[name] = newcg
	return newcg
}
