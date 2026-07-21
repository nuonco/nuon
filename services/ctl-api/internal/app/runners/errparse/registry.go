package errparse

import (
	"sort"
	"sync"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// Registry holds the registered parsers and dispatches a ParseContext to them.
// Parsers are grouped into facet buckets by Tool; within a bucket the signaled
// parsers share one compiled matcher. The bucket layout is built once, lazily,
// on the first Parse so that init-time registration order does not matter.
//
// Registration is expected to happen entirely at init time (parser packages
// blank-imported at the chokepoint). Register panics if called after the first
// Parse, because a late parser would be silently dropped from the already-built
// buckets. That bug is far easier to catch as a panic than as a parser that
// mysteriously never fires.
//
// Parsers are referenced by their index into r.parsers everywhere (buckets,
// dedup, candidates) rather than by interface value: a Parser may be any
// implementation, including a non-comparable struct, so it must never be used
// as a map key.
type Registry struct {
	mu      sync.Mutex
	parsers []Parser
	built   bool

	once     sync.Once
	byTool   map[Tool]*bucket
	agnostic *bucket
	all      *bucket
}

// bucket is the parser set for one facet key (a Tool, the tool-agnostic group,
// or the all-parsers group used on the fail-open path). signaled parsers are
// gated by a shared matcher; always parsers (no signals) are candidates on
// every parse. All entries are global parser indices into Registry.parsers.
type bucket struct {
	signaled []int // global parser indices, gated by matcher
	matcher  *ahoCorasick
	// patternParser maps a matcher pattern id to a position in signaled.
	patternParser []int
	always        []int // global parser indices, always candidates
}

// Register adds a parser. It is intended to be called from init functions and
// panics on misuse (nil parser, empty signal string, or registration after the
// registry has been built).
func (r *Registry) Register(p Parser) {
	if p == nil {
		panic("errparse: Register called with nil parser")
	}
	for _, sig := range p.Signals() {
		if sig == "" {
			panic("errparse: parser declares an empty signal string (use nil signals for an always-candidate parser)")
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.built {
		panic("errparse: Register called after the registry was built (registration must happen at init time)")
	}
	r.parsers = append(r.parsers, p)
}

func (r *Registry) build() {
	r.mu.Lock()
	r.built = true
	parsers := append([]Parser(nil), r.parsers...)
	r.parsers = parsers
	r.mu.Unlock()

	r.byTool = map[Tool]*bucket{}
	r.agnostic = &bucket{}
	r.all = &bucket{}

	for idx, p := range parsers {
		addToBucket(r.all, idx, p)

		tools := p.Tools()
		if len(tools) == 0 {
			addToBucket(r.agnostic, idx, p)
			continue
		}
		for _, t := range tools {
			b := r.byTool[t]
			if b == nil {
				b = &bucket{}
				r.byTool[t] = b
			}
			addToBucket(b, idx, p)
		}
	}

	compileBucket(r.all, parsers)
	compileBucket(r.agnostic, parsers)
	for _, b := range r.byTool {
		compileBucket(b, parsers)
	}
}

func addToBucket(b *bucket, idx int, p Parser) {
	if len(p.Signals()) == 0 {
		b.always = append(b.always, idx)
		return
	}
	b.signaled = append(b.signaled, idx)
}

func compileBucket(b *bucket, parsers []Parser) {
	var patterns []string
	for pos, idx := range b.signaled {
		for _, sig := range parsers[idx].Signals() {
			patterns = append(patterns, sig)
			b.patternParser = append(b.patternParser, pos)
		}
	}
	b.matcher = newAhoCorasick(patterns)
}

// Parse runs the applicable parsers against ctx and returns the first
// confident match in layer order (provider, then tool, then generic), or nil
// when nothing matches. Within a layer, ties are broken by registration order
// so dispatch is deterministic regardless of bucket collection order.
func (r *Registry) Parse(ctx *ParseContext) compositeerrors.CompositeError {
	if ctx == nil || ctx.Raw == "" {
		return nil
	}
	r.once.Do(r.build)

	seen := make([]bool, len(r.parsers))
	var candidates []int

	add := func(idx int) {
		if seen[idx] {
			return
		}
		seen[idx] = true
		candidates = append(candidates, idx)
	}

	collect := func(b *bucket) {
		if b == nil {
			return
		}
		if len(b.signaled) > 0 {
			matched := b.matcher.matchedSet(ctx.Raw)
			for patternID, hit := range matched {
				if hit {
					add(b.signaled[b.patternParser[patternID]])
				}
			}
		}
		for _, idx := range b.always {
			add(idx)
		}
	}

	// Narrow by tool when known; otherwise fail open via the single all-parsers
	// bucket so an undetermined tool stays a one-pass scan instead of scanning
	// every tool bucket separately.
	if ctx.Tool != ToolUnknown {
		collect(r.byTool[ctx.Tool])
		collect(r.agnostic)
	} else {
		collect(r.all)
	}

	// Order candidates by layer (then registration order) up front, then check
	// Applicable and Parse in that order so the most specific parser is
	// consulted first. Evaluating the gates lazily in priority order means a
	// higher-priority match short-circuits every lower-priority parser,
	// including any lazy provider lookup their Applicable would trigger.
	sort.SliceStable(candidates, func(i, j int) bool {
		li, lj := r.parsers[candidates[i]].Layer(), r.parsers[candidates[j]].Layer()
		if li != lj {
			return li < lj
		}
		return candidates[i] < candidates[j]
	})

	for _, idx := range candidates {
		p := r.parsers[idx]
		if !p.Applicable(ctx) {
			continue
		}
		if ce := p.Parse(ctx); ce != nil {
			return ce
		}
	}
	return nil
}

// defaultRegistry is the process-wide registry populated by parser packages'
// init functions (blank-imported at the chokepoint).
var defaultRegistry = &Registry{}

// Register adds a parser to the default registry.
func Register(p Parser) { defaultRegistry.Register(p) }

// Parse dispatches ctx against the default registry.
func Parse(ctx *ParseContext) compositeerrors.CompositeError { return defaultRegistry.Parse(ctx) }
