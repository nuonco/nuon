package config

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const (
	// ConfigPath the path to expose the config on
	ConfigPath = "/config"
)

// Exporter is a config exporter
type Exporter struct {
	httpServer *http.Server
}

// Start starts the config Exporter
func (e *Exporter) Start() error {
	if e.httpServer == nil {
		return nil
	}
	go func() {
		err := e.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			zap.L().Error("failed to start pprof exporter", zap.Error(err))
		}
	}()
	return nil
}

// Stop stops the config exporter
func (e *Exporter) Stop() {
	if e.httpServer == nil {
		return
	}
	_ = e.httpServer.Shutdown(context.Background())
}

// RegisterExporter registers the config exporter
func RegisterExporter(cfg Base, opts ...Option) (*Exporter, error) {

	// Bind the supplied options
	var options options
	for _, opt := range opts {
		opt(&options)
	}

	// Create the http.Server or register the handlers to an existing httpServer
	var httpServer *http.Server
	if options.httpServer == nil {
		httpServer = &http.Server{
			ReadTimeout:  1 * time.Second,
			WriteTimeout: 10 * time.Second,
			Addr:         fmt.Sprintf(":%d", cfg.SystemPort),
		}
		registerRoutes(httpServer)
	} else {
		registerRoutes(options.httpServer)
	}

	exporter := new(Exporter)
	exporter.httpServer = httpServer

	return exporter, nil
}

// registerRoutes registers the config routes
func registerRoutes(httpServer *http.Server) {
	switch obj := httpServer.Handler.(type) {
	case nil:
		config := http.NewServeMux()
		config.HandleFunc(ConfigPath, handler)
		httpServer.Handler = config
	case *http.ServeMux:
		obj.HandleFunc(ConfigPath, handler)
	}
}

// handler is the HTTP handler func
func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	if config := stdLoader.LoadedConfig(); config != nil {
		if _, err := config.WriteTo(w); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}
