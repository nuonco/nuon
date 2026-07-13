package errparse

// ahoCorasick is a compact case-sensitive Aho-Corasick multi-pattern matcher.
// It compiles a set of patterns once and reports, in a single pass over the
// text, which patterns are present. This keeps signal gating O(text length)
// regardless of how many patterns (error signatures) are registered, which is
// what lets the provider layer grow without bound.
//
// It answers only "which patterns are present" (a set), not where, which is all
// the registry needs to decide which parsers are candidates.
type ahoCorasick struct {
	next    []map[byte]int // goto transitions per node
	fail    []int          // failure links
	outputs [][]int        // pattern ids that end at each node
	count   int            // number of patterns
}

// newAhoCorasick builds a matcher over patterns. Pattern ids are their index in
// the slice. Empty patterns are ignored (they would match everywhere).
func newAhoCorasick(patterns []string) *ahoCorasick {
	ac := &ahoCorasick{
		next:    []map[byte]int{{}},
		fail:    []int{0},
		outputs: [][]int{nil},
		count:   len(patterns),
	}

	for id, pat := range patterns {
		if pat == "" {
			continue
		}
		node := 0
		for i := 0; i < len(pat); i++ {
			b := pat[i]
			nxt, ok := ac.next[node][b]
			if !ok {
				nxt = len(ac.next)
				ac.next = append(ac.next, map[byte]int{})
				ac.fail = append(ac.fail, 0)
				ac.outputs = append(ac.outputs, nil)
				ac.next[node][b] = nxt
			}
			node = nxt
		}
		ac.outputs[node] = append(ac.outputs[node], id)
	}

	ac.buildFailureLinks()
	return ac
}

// buildFailureLinks wires the failure links via BFS and folds each node's
// output with its failure target's output so a single lookup at match time
// yields every pattern ending at that position.
func (ac *ahoCorasick) buildFailureLinks() {
	queue := make([]int, 0, len(ac.next))
	for _, child := range ac.next[0] {
		ac.fail[child] = 0
		queue = append(queue, child)
	}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		for b, child := range ac.next[node] {
			queue = append(queue, child)

			f := ac.fail[node]
			for f != 0 {
				if _, ok := ac.next[f][b]; ok {
					break
				}
				f = ac.fail[f]
			}
			if target, ok := ac.next[f][b]; ok && target != child {
				ac.fail[child] = target
			} else {
				ac.fail[child] = 0
			}

			ac.outputs[child] = append(ac.outputs[child], ac.outputs[ac.fail[child]]...)
		}
	}
}

// matchedSet returns a boolean slice of length count where index i is true when
// pattern i occurs in text.
func (ac *ahoCorasick) matchedSet(text string) []bool {
	found := make([]bool, ac.count)
	node := 0
	for i := 0; i < len(text); i++ {
		b := text[i]
		for node != 0 {
			if _, ok := ac.next[node][b]; ok {
				break
			}
			node = ac.fail[node]
		}
		if nxt, ok := ac.next[node][b]; ok {
			node = nxt
		}
		for _, id := range ac.outputs[node] {
			found[id] = true
		}
	}
	return found
}
