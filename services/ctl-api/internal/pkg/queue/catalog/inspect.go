package catalog

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// SignalTypeInfo describes the capabilities and attributes of a registered signal type.
type SignalTypeInfo struct {
	Type             signal.SignalType
	AutoRetry        bool
	MaxRetries       int
	HasCloneSteps    bool
	HasNoOpCheck     bool
	HasPolicyEval    bool
	HasSkipCleanup   bool
	HasOnApprove     bool
	HasOnRetry       bool
	HasOnSkip        bool
	HasOnDeny        bool
	HasFetchSteps    bool
	HasQueue         bool
	Queue            string
	IsParallelizable bool
	HasStepContext   bool
	HasLifecycle     bool
	Operation        string
}

// InspectAll returns information about every registered signal type by instantiating
// each signal and checking which optional interfaces it implements.
func InspectAll() []SignalTypeInfo {
	var infos []SignalTypeInfo
	for typ, constructor := range SignalCatalog {
		sig := constructor()
		infos = append(infos, inspect(typ, sig))
	}
	return infos
}

// InspectType returns information about a single signal type.
func InspectType(typ signal.SignalType) (SignalTypeInfo, error) {
	constructor, ok := SignalCatalog[typ]
	if !ok {
		return SignalTypeInfo{}, fmt.Errorf("signal type %q not registered", typ)
	}
	return inspect(typ, constructor()), nil
}

func inspect(typ signal.SignalType, sig signal.Signal) SignalTypeInfo {
	info := SignalTypeInfo{
		Type:       typ,
		MaxRetries: signal.DefaultMaxRetries,
	}

	if ar, ok := sig.(signal.SignalWithAutoRetry); ok {
		info.AutoRetry = ar.AutoRetry()
	}
	if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
		info.MaxRetries = mr.MaxRetries()
	}
	if _, ok := sig.(signal.SignalWithCloneSteps); ok {
		info.HasCloneSteps = true
	}
	if noop, ok := sig.(signal.SignalWithNoOpCheck); ok {
		info.HasNoOpCheck = noop.IsNoOpCheckable()
	}
	if pe, ok := sig.(signal.SignalWithPolicyEvaluation); ok {
		info.HasPolicyEval = pe.RequiresPolicyEvaluation()
	}
	if _, ok := sig.(signal.SignalWithSkipCleanup); ok {
		info.HasSkipCleanup = true
	}
	if _, ok := sig.(signal.SignalWithOnApprove); ok {
		info.HasOnApprove = true
	}
	if _, ok := sig.(signal.SignalWithOnRetry); ok {
		info.HasOnRetry = true
	}
	if _, ok := sig.(signal.SignalWithOnSkip); ok {
		info.HasOnSkip = true
	}
	if _, ok := sig.(signal.SignalWithOnDeny); ok {
		info.HasOnDeny = true
	}
	if _, ok := sig.(signal.SignalWithFetchSteps); ok {
		info.HasFetchSteps = true
	}
	if q, ok := sig.(signal.SignalWithQueue); ok {
		info.HasQueue = true
		info.Queue = q.Queue()
	}
	if p, ok := sig.(signal.SignalWithParallelizable); ok {
		info.IsParallelizable = p.IsParallelizable()
	}
	if _, ok := sig.(signal.SignalWithStepContext); ok {
		info.HasStepContext = true
	}
	if lc, ok := sig.(signal.SignalWithLifecycleContext); ok {
		info.HasLifecycle = true
		info.Operation = lc.LifecycleContext().Operation
	}

	return info
}
