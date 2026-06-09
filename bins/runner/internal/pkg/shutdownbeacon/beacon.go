// Package shutdownbeacon reports to the control plane when the runner's host VM
// is shutting down.
//
// It subscribes to systemd-logind's PrepareForShutdown D-Bus signal, which
// fires at the very start of an OS shutdown transaction — before networking is
// torn down — on every cloud (AWS, Azure, GCP all deliver an ACPI soft-off to
// the guest on a customer stop/terminate). On that signal the beacon fires a
// bodyless "terminating" beacon at the control plane and force-flushes buffered
// OTEL logs while the network is still up.
//
// It must run in the host (mng) process, not a container, so it can see the
// host system bus. Delivery is best-effort: a hard kill / OOM (SIGKILL) or a
// force-terminate gives no shutdown transaction and therefore no beacon. The
// control plane treats beacon absence via its existing offline-check fallback.
package shutdownbeacon

import (
	"context"
	"os"
	"time"

	"github.com/godbus/dbus/v5"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/process"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

const (
	// beaconTimeout and flushTimeout share the logind inhibitor delay window
	// (InhibitDelayMaxSec, default 5s). Keep their sum under that so logind
	// does not force shutdown to proceed mid-call. The beacon goes first since
	// it is the authoritative signal; the log flush is secondary.
	beaconTimeout = 2 * time.Second
	flushTimeout  = 2 * time.Second
)

type Params struct {
	fx.In

	APIClient   nuonrunner.Client
	Registrar   *process.Registrar
	L           *zap.Logger             `name:"system"`
	LogProvider *otellog.LoggerProvider `name:"process-log-provider" optional:"true"`
	LC          fx.Lifecycle
}

type Beacon struct {
	apiClient   nuonrunner.Client
	registrar   *process.Registrar
	l           *zap.Logger
	logProvider *otellog.LoggerProvider

	conn        *dbus.Conn
	inhibitFile *os.File
	stop        chan struct{}
}

func New(p Params) *Beacon {
	b := &Beacon{
		apiClient:   p.APIClient,
		registrar:   p.Registrar,
		l:           p.L,
		logProvider: p.LogProvider,
		stop:        make(chan struct{}),
	}

	p.LC.Append(fx.Hook{
		OnStart: func(context.Context) error {
			// best-effort: never block runner startup if the system bus or
			// logind is unavailable (e.g. local dev, minimal images).
			b.start()
			return nil
		},
		OnStop: func(context.Context) error {
			b.shutdown()
			return nil
		},
	})

	return b
}

func (b *Beacon) start() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		b.l.Warn("shutdown beacon disabled: no system bus", zap.Error(err))
		return
	}
	b.conn = conn

	b.takeInhibitor()

	if err := conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.login1.Manager"),
		dbus.WithMatchMember("PrepareForShutdown"),
	); err != nil {
		b.l.Warn("shutdown beacon disabled: cannot match PrepareForShutdown", zap.Error(err))
		return
	}

	ch := make(chan *dbus.Signal, 1)
	conn.Signal(ch)
	go b.loop(ch)

	b.l.Info("shutdown beacon armed")
}

// takeInhibitor acquires a "delay" inhibitor lock so logind waits (up to
// InhibitDelayMaxSec) after emitting PrepareForShutdown before proceeding,
// guaranteeing a window for the beacon + log flush. Releasing the lock (closing
// the fd) lets shutdown continue.
func (b *Beacon) takeInhibitor() {
	var fd dbus.UnixFD
	mgr := b.conn.Object("org.freedesktop.login1", "/org/freedesktop/login1")
	if err := mgr.Call(
		"org.freedesktop.login1.Manager.Inhibit", 0,
		"shutdown", "nuon-runner", "report VM termination to control plane", "delay",
	).Store(&fd); err != nil {
		b.l.Warn("could not take shutdown inhibitor", zap.Error(err))
		return
	}
	b.inhibitFile = os.NewFile(uintptr(fd), "shutdown-inhibitor")
}

func (b *Beacon) loop(ch <-chan *dbus.Signal) {
	for {
		select {
		case <-b.stop:
			return
		case sig := <-ch:
			if sig == nil || len(sig.Body) != 1 {
				continue
			}
			down, ok := sig.Body[0].(bool)
			if !ok || !down {
				// false => a previously scheduled shutdown was cancelled.
				continue
			}
			b.onShutdown()
			return
		}
	}
}

func (b *Beacon) onShutdown() {
	b.l.Info("host shutdown detected, firing terminating beacon")
	b.fire()
	b.flushLogs()
	b.releaseInhibitor()
}

func (b *Beacon) fire() {
	ctx, cancel := context.WithTimeout(context.Background(), beaconTimeout)
	defer cancel()

	if err := b.apiClient.ReportTerminating(ctx, b.registrar.ProcessID()); err != nil {
		// best-effort breadcrumb only; the authoritative "did it arrive" signal
		// is server-side. This log races the same network teardown.
		b.l.Warn("terminating beacon failed", zap.Error(err))
	}
}

func (b *Beacon) flushLogs() {
	if b.logProvider == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	defer cancel()
	if err := b.logProvider.ForceFlush(ctx); err != nil {
		b.l.Warn("log force-flush during shutdown failed", zap.Error(err))
	}
}

func (b *Beacon) releaseInhibitor() {
	if b.inhibitFile != nil {
		_ = b.inhibitFile.Close()
		b.inhibitFile = nil
	}
}

func (b *Beacon) shutdown() {
	close(b.stop)
	b.releaseInhibitor()
	if b.conn != nil {
		_ = b.conn.Close()
	}
}
