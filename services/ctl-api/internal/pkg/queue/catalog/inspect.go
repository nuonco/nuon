package catalog

import (
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// SignalTypeInfo describes the capabilities and attributes of a registered signal type.
type SignalTypeInfo struct {
	Type             signal.SignalType
	Namespace        string
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

// deriveNamespace extracts a namespace from a signal type string.
// e.g., "install-created" -> "install", "component-deploy-apply-plan" -> "component"
func deriveNamespace(typ signal.SignalType) string {
	s := string(typ)
	if idx := strings.Index(s, "-"); idx > 0 {
		return s[:idx]
	}
	return s
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

// safeCall runs fn and recovers from any panic, returning the zero value on failure.
func safeCall[T any](fn func() T) (result T) {
	defer func() {
		if r := recover(); r != nil {
			// Method panicked on zero-value receiver; use zero value.
		}
	}()
	return fn()
}

func inspect(typ signal.SignalType, sig signal.Signal) SignalTypeInfo {
	info := SignalTypeInfo{
		Type:       typ,
		Namespace:  deriveNamespace(typ),
		MaxRetries: signal.DefaultMaxRetries,
	}

	if ar, ok := sig.(signal.SignalWithAutoRetry); ok {
		info.AutoRetry = safeCall(func() bool { return ar.AutoRetry() })
	}
	if mr, ok := sig.(signal.SignalWithMaxRetries); ok {
		info.MaxRetries = safeCall(func() int { return mr.MaxRetries() })
	}
	if _, ok := sig.(signal.SignalWithCloneSteps); ok {
		info.HasCloneSteps = true
	}
	if noop, ok := sig.(signal.SignalWithNoOpCheck); ok {
		info.HasNoOpCheck = safeCall(func() bool { return noop.IsNoOpCheckable() })
	}
	if pe, ok := sig.(signal.SignalWithPolicyEvaluation); ok {
		info.HasPolicyEval = safeCall(func() bool { return pe.RequiresPolicyEvaluation() })
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
		info.Queue = safeCall(func() string { return q.Queue() })
	}
	if p, ok := sig.(signal.SignalWithParallelizable); ok {
		info.IsParallelizable = safeCall(func() bool { return p.IsParallelizable() })
	}
	if _, ok := sig.(signal.SignalWithStepContext); ok {
		info.HasStepContext = true
	}
	if lc, ok := sig.(signal.SignalWithLifecycleContext); ok {
		info.HasLifecycle = true
		info.Operation = safeCall(func() string { return lc.LifecycleContext().Operation })
	}

	return info
}
