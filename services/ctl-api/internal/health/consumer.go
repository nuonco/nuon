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

	HB  *runnersconsumer.HeartbeatConsumer
	OL  *runnersconsumer.OtelLogsConsumer
	OT  *runnersconsumer.OtelTracesConsumer
	DLQ *consumer.DLQConsumer
}

type ConsumerHealthcheckServer struct {
	cfg *internal.Config
	l   *zap.Logger
	srv *http.Server

	maxStuck time.Duration
	sinks    []healthySink
}

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
