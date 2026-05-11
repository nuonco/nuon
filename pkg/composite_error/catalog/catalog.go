// Package catalog is the in-memory registry for CompositeError types and
// Parsers, populated via init() blocks in per-type packages.
//
// This mirrors pkg/queue/catalog: there is no DB-backed catalog table.
// On every process start, importing a type's package runs its init() which
// registers the factory and parser(s) here.
package catalog

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
)

var (
	typeMu       sync.RWMutex
	typeRegistry = map[composite_error.Type]func() composite_error.CompositeError{}

	parserMu     sync.RWMutex
	parsersByCtx = map[composite_error.ParseContext][]composite_error.Parser{}
	allParsers   []composite_error.Parser
)

// RegisterType associates a CompositeError type with its factory. Panics on
// duplicate registration — same contract as pkg/queue/catalog.
//
// The factory must return a fresh, zero-valued instance of the implementing
// struct on each call.
func RegisterType(typ composite_error.Type, fn func() composite_error.CompositeError) {
	typeMu.Lock()
	defer typeMu.Unlock()

	if _, exists := typeRegistry[typ]; exists {
		panic(fmt.Sprintf("composite_error: duplicate type registered: %q", typ))
	}
	typeRegistry[typ] = fn
}

// LookupType returns the factory registered for typ, or (nil, false).
func LookupType(typ composite_error.Type) (func() composite_error.CompositeError, bool) {
	typeMu.RLock()
	defer typeMu.RUnlock()

	fn, ok := typeRegistry[typ]
	return fn, ok
}

// Hydrate looks up the factory for typ, instantiates an empty value, and
// unmarshals data into it. Returns an error if the type is not registered or
// the JSON does not match the type's schema.
func Hydrate(typ composite_error.Type, data []byte) (composite_error.CompositeError, error) {
	fn, ok := LookupType(typ)
	if !ok {
		return nil, fmt.Errorf("composite_error: unknown type %q", typ)
	}

	val := fn()
	if len(data) == 0 {
		return val, nil
	}
	if err := json.Unmarshal(data, val); err != nil {
		return nil, fmt.Errorf("composite_error: hydrate %q: %w", typ, err)
	}
	return val, nil
}

// Types returns all registered type identifiers, sorted. Useful for the
// admin dashboard catalog browser.
func Types() []composite_error.Type {
	typeMu.RLock()
	defer typeMu.RUnlock()

	out := make([]composite_error.Type, 0, len(typeRegistry))
	for t := range typeRegistry {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// RegisterParser indexes p under each of its declared ParseContexts. Panics
// if Contexts() is empty — every parser must opt in to at least one subtree
// (see specs/composite-errors.md).
func RegisterParser(p composite_error.Parser) {
	parserMu.Lock()
	defer parserMu.Unlock()

	ctxs := p.Contexts()
	if len(ctxs) == 0 {
		panic(fmt.Sprintf("composite_error: parser %q registered with no contexts", p.Name()))
	}
	for _, c := range ctxs {
		parsersByCtx[c] = append(parsersByCtx[c], p)
	}
	allParsers = append(allParsers, p)
}

// ParsersForContext returns the parsers that should be invoked for a given
// dispatch context, ordered most-specific to least-specific. Within a level
// parsers are returned in registration order.
//
// Walking is by "/"-segment ancestry: "terraform/plan" matches parsers
// registered against "terraform/plan", "terraform", and "" (root).
func ParsersForContext(c composite_error.ParseContext) []composite_error.Parser {
	parserMu.RLock()
	defer parserMu.RUnlock()

	ancestors := ancestorsOf(c)
	out := make([]composite_error.Parser, 0)
	for _, a := range ancestors {
		out = append(out, parsersByCtx[a]...)
	}
	return out
}

// AllParsers returns every registered parser, in registration order.
// Used by the admin dashboard catalog browser.
func AllParsers() []composite_error.Parser {
	parserMu.RLock()
	defer parserMu.RUnlock()

	out := make([]composite_error.Parser, len(allParsers))
	copy(out, allParsers)
	return out
}

// ancestorsOf returns c, its parents, and "" (root), most-specific first.
//
//	"terraform/plan/init" → ["terraform/plan/init", "terraform/plan", "terraform", ""]
//	"terraform"           → ["terraform", ""]
//	""                    → [""]
func ancestorsOf(c composite_error.ParseContext) []composite_error.ParseContext {
	s := string(c)
	out := []composite_error.ParseContext{c}
	for {
		i := lastSlash(s)
		if i < 0 {
			break
		}
		s = s[:i]
		out = append(out, composite_error.ParseContext(s))
	}
	if c != "" {
		out = append(out, "")
	}
	return out
}

func lastSlash(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '/' {
			return i
		}
	}
	return -1
}

// Reset clears all registrations. Test-only.
func Reset() {
	typeMu.Lock()
	parserMu.Lock()
	defer typeMu.Unlock()
	defer parserMu.Unlock()

	typeRegistry = map[composite_error.Type]func() composite_error.CompositeError{}
	parsersByCtx = map[composite_error.ParseContext][]composite_error.Parser{}
	allParsers = nil
}
