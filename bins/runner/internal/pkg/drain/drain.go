package drain

import (
	"fmt"
	"sync"
	"time"
)

type Drainer struct {
	drainOnce sync.Once
	drainCh   chan struct{}

	mu      sync.Mutex
	doneChs []<-chan struct{}
}

func New() *Drainer {
	return &Drainer{
		drainCh: make(chan struct{}),
	}
}

func (d *Drainer) Register(ch <-chan struct{}) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.doneChs = append(d.doneChs, ch)
}

func (d *Drainer) Drain() {
	d.drainOnce.Do(func() {
		close(d.drainCh)
	})
}

func (d *Drainer) DrainCh() <-chan struct{} {
	return d.drainCh
}

func (d *Drainer) IsDraining() bool {
	select {
	case <-d.drainCh:
		return true
	default:
		return false
	}
}

func (d *Drainer) Wait(timeout time.Duration) error {
	d.mu.Lock()
	chs := make([]<-chan struct{}, len(d.doneChs))
	copy(chs, d.doneChs)
	d.mu.Unlock()

	if len(chs) == 0 {
		return nil
	}

	deadline := time.After(timeout)
	for _, ch := range chs {
		select {
		case <-ch:
		case <-deadline:
			return fmt.Errorf("drain timed out after %s", timeout)
		}
	}
	return nil
}
