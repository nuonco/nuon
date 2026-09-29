package errparse

import (
	"sort"
	"sync"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type Registry struct {
	mu      sync.Mutex
	parsers []Parser
	built   bool

	once     sync.Once
	byTool   map[Tool]*bucket
	agnostic *bucket
	all      *bucket
}

type bucket struct {
	signaled      []int
	matcher       *ahoCorasick
	patternParser []int
	always        []int
}

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

	// why: Narrow by tool when known; otherwise fail open via the single all-parsers
	// bucket so an undetermined tool stays a one-pass scan instead of scanning
	// every tool bucket separately.
	if ctx.Tool != ToolUnknown {
		collect(r.byTool[ctx.Tool])
		collect(r.agnostic)
	} else {
		collect(r.all)
	}

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

var defaultRegistry = &Registry{}

func Register(p Parser) { defaultRegistry.Register(p) }

func Parse(ctx *ParseContext) compositeerrors.CompositeError { return defaultRegistry.Parse(ctx) }
