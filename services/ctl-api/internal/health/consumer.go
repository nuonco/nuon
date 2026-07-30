package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	runnersconsumer "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/consumer"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/consumer"
)

type ConsumerHealthcheckParams struct {
	fx.In

	Cfg *internal.Config
	L   *zap.Logger

	// Each is nil when that consumer isn't selected/enabled in this process
	// (see consumer.NewSink) — Healthy on a nil sink always reports healthy.
	HB  *runnersconsumer.HeartbeatConsumer
	OL  *runnersconsumer.OtelLogsConsumer
	OT  *runnersconsumer.OtelTracesConsumer
	DLQ *consumer.DLQConsumer
}

// ConsumerHealthcheckServer exposes /livez for the `consumer` command.
//
// Deliberately narrower than WorkerHealthcheckServer: it checks only whether
// a handler call is stuck (pkgkafka.Consumer.Stuck), never ClickHouse or
// Kafka reachability directly. Restarting a consumer pod doesn't fix a
// degraded dependency, and pinging one from every replica risks a
// synchronized restart storm exactly when the dependency is already
// struggling. This only fires for something a restart can actually fix: an
// in-process hang the write timeout itself failed to bound.
//
// Always started — no enabled flag. The only case you'd want it off (running
// two named consumer processes by hand on one host) is solved by pointing one
// at a different consumer_healthcheck_port, not a toggle.
type ConsumerHealthcheckServer struct {
	cfg *internal.Config
	l   *zap.Logger
	srv *http.Server

	maxStuck time.Duration
	sinks    []healthySink
}

// healthySink is the subset of consumer.Sink — which every consumer embeds —
// that this server needs.
type healthySink interface {
	ConsumerName() string
	Healthy(max time.Duration) (bool, time.Duration)
}

func NewConsumerHealthcheck(params ConsumerHealthcheckParams) *ConsumerHealthcheckServer {
	var sinks []healthySink
	if params.HB != nil {
		sinks = append(sinks, params.HB)
	}
	if params.OL != nil {
		sinks = append(sinks, params.OL)
	}
	if params.OT != nil {
		sinks = append(sinks, params.OT)
	}
	if params.DLQ != nil {
		sinks = append(sinks, params.DLQ)
	}

	return &ConsumerHealthcheckServer{
		cfg:      params.Cfg,
		l:        params.L.Named("consumer-healthcheck"),
		maxStuck: params.Cfg.KafkaConsumerLivenessTimeout,
		sinks:    sinks,
	}
}

func (h *ConsumerHealthcheckServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/livez", h.livezHandler)

	addr := fmt.Sprintf("0.0.0.0:%s", h.cfg.ConsumerHealthcheckPort)
	h.srv = &http.Server{Addr: addr, Handler: mux}

	h.l.Info("starting consumer healthcheck server", zap.String("addr", addr))
	go func() {
		if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			h.l.Error("consumer healthcheck server error", zap.Error(err))
		}
	}()
	return nil
}

func (h *ConsumerHealthcheckServer) Stop(ctx context.Context) error {
	if h.srv == nil {
		return nil
	}
	h.l.Info("stopping consumer healthcheck server")
	return h.srv.Shutdown(ctx)
}

// livezHandler fails the pod if any consumer it hosts is stuck. A deployment can
// host several (the `otel` deployment runs otel-logs and otel-traces), so the
// response names the stuck ones: otherwise a restart tells you a pod wedged but
// not which consumer caused it, and the pod that gets killed is also carrying
// the healthy one's partitions.
func (h *ConsumerHealthcheckServer) livezHandler(rw http.ResponseWriter, _ *http.Request) {
	var stuckFor time.Duration
	var stuckNames []string
	for _, s := range h.sinks {
		ok, d := s.Healthy(h.maxStuck)
		if !ok {
			stuckNames = append(stuckNames, s.ConsumerName())
			if d > stuckFor {
				stuckFor = d
			}
		}
	}

	if len(stuckNames) > 0 {
		writeJSON(rw, http.StatusServiceUnavailable, map[string]any{
			"status":       "error",
			"stuck":        stuckNames,
			"stuck_for_ms": stuckFor.Milliseconds(),
		})
		return
	}

	writeJSON(rw, http.StatusOK, map[string]any{
		"status": "ok",
	})
}
