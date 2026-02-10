package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	temporalsdk "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporalclient "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

type WorkerHealthcheckParams struct {
	fx.In

	Cfg     *internal.Config
	DB      *gorm.DB `name:"psql"`
	CHDB    *gorm.DB `name:"ch"`
	TClient temporalclient.Client
	MW      metrics.Writer
	L       *zap.Logger
}

type WorkerHealthcheckServer struct {
	cfg     *internal.Config
	db      *gorm.DB
	chDB    *gorm.DB
	tclient temporalclient.Client
	mw      metrics.Writer
	l       *zap.Logger
	srv     *http.Server
}

func NewWorkerHealthcheck(params WorkerHealthcheckParams) *WorkerHealthcheckServer {
	return &WorkerHealthcheckServer{
		cfg:     params.Cfg,
		db:      params.DB,
		chDB:    params.CHDB,
		tclient: params.TClient,
		mw:      params.MW,
		l:       params.L,
	}
}

func (w *WorkerHealthcheckServer) Start() error {
	if !w.cfg.WorkerHealthcheckEnabled {
		w.l.Info("worker healthcheck server disabled")
		return nil
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/livez", w.livezHandler)
	mux.HandleFunc("/readyz", w.readyzHandler)
	mux.HandleFunc("/version", w.versionHandler)

	addr := fmt.Sprintf("0.0.0.0:%s", w.cfg.WorkerHealthcheckPort)
	w.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	w.l.Info("starting worker healthcheck server", zap.String("addr", addr))
	go func() {
		if err := w.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			w.l.Error("worker healthcheck server error", zap.Error(err))
		}
	}()
	return nil
}

func (w *WorkerHealthcheckServer) Stop(ctx context.Context) error {
	if w.srv == nil {
		return nil
	}
	w.l.Info("stopping worker healthcheck server")
	return w.srv.Shutdown(ctx)
}

func (w *WorkerHealthcheckServer) livezHandler(rw http.ResponseWriter, r *http.Request) {
	sqlDB, err := w.db.DB()
	if err != nil {
		writeJSON(rw, http.StatusServiceUnavailable, map[string]any{
			"status": "error",
			"error":  "unable to get psql connection",
		})
		return
	}
	if err := sqlDB.PingContext(r.Context()); err != nil {
		w.mw.Incr("worker.healthcheck.check", metrics.ToTags(map[string]string{
			"system": "psql",
			"status": "unable_to_ping",
		}))
		writeJSON(rw, http.StatusServiceUnavailable, map[string]any{
			"status": "error",
			"error":  "unable to ping psql db",
		})
		return
	}
	w.mw.Incr("worker.healthcheck.check", metrics.ToTags(map[string]string{
		"system": "psql",
		"status": "ok",
	}))

	degraded := make([]string, 0)

	chDB, err := w.chDB.DB()
	if err != nil {
		degraded = append(degraded, "ch")
		w.mw.Incr("worker.healthcheck.check", metrics.ToTags(map[string]string{
			"system": "ch",
			"status": "unable_to_connect",
		}))
	} else {
		if err := chDB.PingContext(r.Context()); err != nil {
			degraded = append(degraded, "ch")
			w.mw.Incr("worker.healthcheck.check", metrics.ToTags(map[string]string{
				"system": "ch",
				"status": "unable_to_ping",
			}))
		} else {
			w.mw.Incr("worker.healthcheck.check", metrics.ToTags(map[string]string{
				"system": "ch",
				"status": "ok",
			}))
		}
	}

	_, err = w.tclient.CheckHealth(r.Context(), &temporalsdk.CheckHealthRequest{})
	if err != nil {
		degraded = append(degraded, "temporal")
		w.mw.Incr("worker.healthcheck.check", metrics.ToTags(map[string]string{
			"system": "temporal",
			"status": "unable_to_ping",
		}))
	} else {
		w.mw.Incr("worker.healthcheck.check", metrics.ToTags(map[string]string{
			"system": "temporal",
			"status": "ok",
		}))
	}

	statusCode := http.StatusOK
	status := "ok"
	if len(degraded) > 0 {
		status = "degraded"
		statusCode = http.StatusMultiStatus
	}

	writeJSON(rw, statusCode, map[string]any{
		"status":   status,
		"degraded": degraded,
	})
}

func (w *WorkerHealthcheckServer) readyzHandler(rw http.ResponseWriter, _ *http.Request) {
	writeJSON(rw, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func (w *WorkerHealthcheckServer) versionHandler(rw http.ResponseWriter, _ *http.Request) {
	writeJSON(rw, http.StatusOK, map[string]any{
		"version": w.cfg.Version,
		"git_ref": w.cfg.GitRef,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
