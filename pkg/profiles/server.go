package profiles

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"go.uber.org/fx"
)

type ProfilerOptions struct {
	Enabled bool
	Port    int
}

// DefaultProfilerOptions provides default settings for the profiler
func DefaultProfilerOptions() ProfilerOptions {
	return ProfilerOptions{
		Enabled: true,
		Port:    6060,
	}
}

// NewProfilerServer creates an HTTP server for profiling
func NewProfilerServer(options ProfilerOptions) *http.Server {
	if !options.Enabled {
		return nil
	}

	return &http.Server{
		Handler: SetupPprofMux(),
		Addr:    fmt.Sprintf(":%d", options.Port),
	}
}

// RegisterProfiler registers the profiler with fx lifecycle
func RegisterProfiler(lc fx.Lifecycle, options ProfilerOptions) {
	if !options.Enabled {
		return
	}

	server := NewProfilerServer(options)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return err
			}

			go func() {
				fmt.Printf("Starting pprof server on http://localhost%s/debug/pprof/\n", server.Addr)
				if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
					fmt.Printf("Error starting pprof server: %v\n", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return server.Shutdown(ctx)
		},
	})
}
