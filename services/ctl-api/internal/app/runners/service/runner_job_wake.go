package service

import (
	"hash/fnv"
	"sync"
)

const runnerJobWakeShards = 32

type runnerJobWakeShard struct {
	mu      sync.Mutex
	waiters map[string]map[chan struct{}]struct{}
}

type RunnerJobWakeRegistry struct {
	shards [runnerJobWakeShards]*runnerJobWakeShard
}

func NewRunnerJobWakeRegistry() *RunnerJobWakeRegistry {
	r := &RunnerJobWakeRegistry{}
	for i := range r.shards {
		r.shards[i] = &runnerJobWakeShard{
			waiters: make(map[string]map[chan struct{}]struct{}),
		}
	}
	return r
}

func (r *RunnerJobWakeRegistry) shardFor(runnerID string) *runnerJobWakeShard {
	h := fnv.New32a()
	_, _ = h.Write([]byte(runnerID))
	return r.shards[h.Sum32()%runnerJobWakeShards]
}

// why: Subscribe registers a waiter for a runner's job-available wakeups. The
// returned channel is buffered (cap 1) and coalescing: Wake does a non-blocking
// send, so a waiter that hasn't drained its previous signal isn't blocked on.
// Callers MUST invoke the returned unsubscribe func (defer) to avoid leaking
// the waiter.
func (r *RunnerJobWakeRegistry) Subscribe(runnerID string) (<-chan struct{}, func()) {
	ch := make(chan struct{}, 1)
	sh := r.shardFor(runnerID)

	sh.mu.Lock()
	set := sh.waiters[runnerID]
	if set == nil {
		set = make(map[chan struct{}]struct{})
		sh.waiters[runnerID] = set
	}
	set[ch] = struct{}{}
	sh.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			sh.mu.Lock()
			if cur := sh.waiters[runnerID]; cur != nil {
				delete(cur, ch)
				if len(cur) == 0 {
					delete(sh.waiters, runnerID)
				}
			}
			sh.mu.Unlock()
		})
	}

	return ch, unsubscribe
}

func (r *RunnerJobWakeRegistry) Wake(runnerID string) int {
	sh := r.shardFor(runnerID)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	set := sh.waiters[runnerID]
	for ch := range set {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
	return len(set)
}
