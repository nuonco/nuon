package errparse

type ahoCorasick struct {
	next    []map[byte]int
	fail    []int
	outputs [][]int
	count   int
}

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
