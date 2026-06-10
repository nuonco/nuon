package service

import (
	"hash/fnv"
	"sync"
)

// runnerJobWakeShards bounds lock contention on the registry: Subscribe/Wake
// take only their runner's shard lock, so a burst of NOTIFY-driven wakes for
// different runners doesn't serialize on a single global mutex.
const runnerJobWakeShards = 32

type runnerJobWakeShard struct {
	mu      sync.Mutex
	waiters map[string]map[chan struct{}]struct{}
}

// RunnerJobWakeRegistry fans a single pod-level NOTIFY listener out to the
// parked TailRunnerJobs handlers on this pod, keyed by runner_id. The wake
// carries no job data — it only tells a parked handler to re-probe Postgres —
// so a missed or spurious wake is harmless (the handler's poll backstop and
// authoritative re-query cover correctness).
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

// Subscribe registers a waiter for a runner's job-available wakeups. The
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

// Wake signals every current waiter for a runner to re-probe and returns how
// many waiters were signalled (0 means no parked request for this runner lives
// on this pod). Non-blocking and coalescing.
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
