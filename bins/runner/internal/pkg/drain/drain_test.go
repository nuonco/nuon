package drain

import (
	"testing"
	"time"
)

func TestDrainer_DrainClosesChannel(t *testing.T) {
	d := New()

	select {
	case <-d.DrainCh():
		t.Fatal("drain channel should not be closed before Drain()")
	default:
	}

	if d.IsDraining() {
		t.Fatal("should not be draining before Drain()")
	}

	d.Drain()

	select {
	case <-d.DrainCh():
	default:
		t.Fatal("drain channel should be closed after Drain()")
	}

	if !d.IsDraining() {
		t.Fatal("should be draining after Drain()")
	}
}

func TestDrainer_DrainIsIdempotent(t *testing.T) {
	d := New()
	d.Drain()
	d.Drain()

	if !d.IsDraining() {
		t.Fatal("should still be draining after double Drain()")
	}
}

func TestDrainer_WaitNoDoneChannels(t *testing.T) {
	d := New()
	d.Drain()

	if err := d.Wait(time.Second); err != nil {
		t.Fatalf("Wait with no registered channels should return nil: %v", err)
	}
}

func TestDrainer_WaitAllDone(t *testing.T) {
	d := New()

	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	d.Register(ch1)
	d.Register(ch2)

	d.Drain()

	go func() {
		time.Sleep(10 * time.Millisecond)
		close(ch1)
		time.Sleep(10 * time.Millisecond)
		close(ch2)
	}()

	if err := d.Wait(time.Second); err != nil {
		t.Fatalf("Wait should succeed when all channels close: %v", err)
	}
}

func TestDrainer_WaitTimeout(t *testing.T) {
	d := New()

	ch := make(chan struct{})
	d.Register(ch)

	d.Drain()

	err := d.Wait(50 * time.Millisecond)
	if err == nil {
		t.Fatal("Wait should return error on timeout")
	}
}

func TestDrainer_WaitPartialDone(t *testing.T) {
	d := New()

	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	d.Register(ch1)
	d.Register(ch2)

	d.Drain()

	close(ch1)

	err := d.Wait(50 * time.Millisecond)
	if err == nil {
		t.Fatal("Wait should timeout when only some channels close")
	}
}
