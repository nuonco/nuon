package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/powertoolsdev/api/internal"
	statusserver "github.com/powertoolsdev/api/internal/servers/status"
	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/protos/api/generated/types/status/v1/statusv1connect"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var runServerCmd = &cobra.Command{
	Use:   "server",
	Short: "run server",
	Run:   runServer,
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(runServerCmd)
}

func registerLoadbalancerHealthCheck(mux *http.ServeMux) {
	mux.Handle("/_ping", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusOK)

		if _, err := rw.Write([]byte("{\"status\": \"ok\"}")); err != nil {
			log.Fatal("unable to write load balancer health check response", err.Error())
		}
	}))
}

// registerStatusServer registers the status service handler on the provided mux
func registerStatusServer(mux *http.ServeMux, cfg *internal.Config) error {
	srv, err := statusserver.New(statusserver.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to initialize status server: %w", err)
	}

	path, handler := statusv1connect.NewStatusServiceHandler(srv)
	mux.Handle(path, handler)
	return nil
}

//nolint:all
func runServer(cmd *cobra.Command, args []string) {
	var cfg internal.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		log.Fatalf("failed to load config: %s", err)
	}

	var (
		l   *zap.Logger
		err error
	)
	switch cfg.Env {
	case config.Development:
		l, err = zap.NewDevelopment()
	default:
		l, err = zap.NewProduction()
	}
	zap.ReplaceGlobals(l)
	if err != nil {
		fmt.Printf("failed to instantiate logger: %v\n", err)
	}

	mux := http.NewServeMux()
	if err := registerStatusServer(mux, &cfg); err != nil {
		l.Fatal("unable to register status server:", zap.Error(err))
	}
	registerLoadbalancerHealthCheck(mux)

	l.Info("server starting: ",
		zap.String("host", cfg.HTTPAddress),
		zap.String("port", cfg.HTTPPort))

	if err := http.ListenAndServe(
		fmt.Sprintf("%s:%s", cfg.HTTPAddress, cfg.HTTPPort),

		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		l.Fatal("error on listen and server", zap.Error(err))
	}
}
