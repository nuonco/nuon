package audit

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
	"go.uber.org/zap"

	runnerconfig "github.com/nuonco/nuon/pkg/runner/config"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

type recordingClientWriter struct {
	enableCalls   atomic.Int32
	disableCalls  atomic.Int32
	stoppingCalls atomic.Int32
	failures      int32
	enabled       chan struct{}
	disabled      chan struct{}
}

func (w *recordingClientWriter) Enable() error {
	if w.enableCalls.Add(1) <= w.failures {
		return errors.New("route rejected startup event")
	}
	select {
	case w.enabled <- struct{}{}:
	default:
	}
	return nil
}

func (w *recordingClientWriter) Disable() {
	w.disableCalls.Add(1)
	select {
	case w.disabled <- struct{}{}:
	default:
	}
}

func (w *recordingClientWriter) ProcessStopping(context.Context, string, string) error {
	w.stoppingCalls.Add(1)
	return nil
}

func TestClientWaitsForRouteAndRetriesStartupDelivery(t *testing.T) {
	w := &recordingClientWriter{failures: 1, enabled: make(chan struct{}, 1), disabled: make(chan struct{}, 1)}
	var availabilityChecks atomic.Int32
	c := &Client{
		installID: "inst-test",
		logger:    zap.NewNop(),
		writer:    w,
		done:      make(chan struct{}),
		routeAvailableFn: func() bool {
			return availabilityChecks.Add(1) > 1
		},
		pollInterval:  time.Millisecond,
		retryInterval: time.Millisecond,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go c.run(ctx)

	select {
	case <-w.enabled:
	case <-time.After(time.Second):
		t.Fatal("client did not enable audit delivery")
	}
	cancel()
	select {
	case <-c.done:
	case <-time.After(time.Second):
		t.Fatal("client did not stop")
	}
	if got := w.enableCalls.Load(); got != 2 {
		t.Fatalf("startup delivery attempts = %d, want 2", got)
	}
	if got := availabilityChecks.Load(); got < 3 {
		t.Fatalf("route availability checks = %d, want at least 3", got)
	}
}

func TestClientDisablesDeliveryWhenRouteDisappears(t *testing.T) {
	w := &recordingClientWriter{enabled: make(chan struct{}, 1), disabled: make(chan struct{}, 1)}
	var available atomic.Bool
	available.Store(true)
	c := &Client{
		installID:        "inst-test",
		logger:           zap.NewNop(),
		writer:           w,
		done:             make(chan struct{}),
		routeAvailableFn: available.Load,
		pollInterval:     time.Millisecond,
		retryInterval:    time.Millisecond,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go c.run(ctx)

	select {
	case <-w.enabled:
	case <-time.After(time.Second):
		t.Fatal("client did not enable audit delivery")
	}
	available.Store(false)
	select {
	case <-w.disabled:
	case <-time.After(time.Second):
		t.Fatal("client did not disable audit delivery after the route disappeared")
	}
	cancel()
	select {
	case <-c.done:
	case <-time.After(time.Second):
		t.Fatal("client did not stop")
	}
}

func TestClientStopWritesAuditLifecycleEvent(t *testing.T) {
	w := &recordingClientWriter{enabled: make(chan struct{}, 1), disabled: make(chan struct{}, 1)}
	c := &Client{installID: "inst-test", logger: zap.NewNop(), writer: w}

	if err := c.stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := w.stoppingCalls.Load(); got != 1 {
		t.Fatalf("stopping lifecycle events = %d, want 1", got)
	}
}

type orderedExporter struct {
	mu    *sync.Mutex
	order *[]string
}

func (e *orderedExporter) Export(context.Context, []sdklog.Record) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	*e.order = append(*e.order, "audit")
	return nil
}

func (*orderedExporter) Shutdown(context.Context) error {
	return nil
}

type testRouteLifecycle struct{}

func (testRouteLifecycle) AuditRouteLifecycle() {}

func TestClientStopsBeforeOwnedRouteLifecycle(t *testing.T) {
	var order []string
	var mu sync.Mutex
	exporter := &orderedExporter{mu: &mu, order: &order}
	w := newWriter(nil, nil, new(testExporter), exporter, nil)
	w.enabled = true

	app := fx.New(
		fx.Supply(w),
		fx.Supply(&settings.Settings{Cfg: &runnerconfig.Config{IsNuonctl: true}, Metadata: map[string]string{"install.id": "inst-test"}}),
		fx.Supply(fx.Annotate(zap.NewNop(), fx.ResultTags(`name:"system"`))),
		fx.Provide(func(lifecycle fx.Lifecycle) LocalRouteLifecycle {
			lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
				mu.Lock()
				defer mu.Unlock()
				order = append(order, "route")
				return nil
			}})
			return testRouteLifecycle{}
		}),
		fx.Provide(NewClient),
		fx.Invoke(func(*Client) {}),
		fx.NopLogger,
	)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := app.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if err := app.Stop(ctx); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 || order[0] != "audit" || order[1] != "route" {
		t.Fatalf("shutdown order = %q, want audit before route", order)
	}
}
