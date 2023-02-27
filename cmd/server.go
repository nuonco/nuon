package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/orgs-api/internal"
	"github.com/powertoolsdev/orgs-api/internal/orgcontext"
	"github.com/powertoolsdev/orgs-api/internal/servers"
	appsserver "github.com/powertoolsdev/orgs-api/internal/servers/apps"
	deploymentsserver "github.com/powertoolsdev/orgs-api/internal/servers/deployments"
	installserver "github.com/powertoolsdev/orgs-api/internal/servers/installs"
	instancesserver "github.com/powertoolsdev/orgs-api/internal/servers/instances"
	orgsserver "github.com/powertoolsdev/orgs-api/internal/servers/orgs"
	statusserver "github.com/powertoolsdev/orgs-api/internal/servers/status"
	"github.com/powertoolsdev/protos/shared/generated/types/status/v1/statusv1connect"
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

// registerAppsServer registers the apps service handler on the provided mux
func registerAppsServer(v *validator.Validate, mux *http.ServeMux, cfg *internal.Config) error {
	ctxProvider, err := orgcontext.NewStaticProvider(v, orgcontext.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create orgcontext provider: %w", err)
	}

	path, handler, err := appsserver.NewHandler(v, servers.WithContextProvider(ctxProvider))
	if err != nil {
		return fmt.Errorf("unable to initialize status server: %w", err)
	}

	mux.Handle(path, handler)
	return nil
}

// registerInstancesServer registers the installs service handler on the provided mux
func registerInstancesServer(v *validator.Validate, mux *http.ServeMux, cfg *internal.Config) error {
	ctxProvider, err := orgcontext.NewStaticProvider(v, orgcontext.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create orgcontext provider: %w", err)
	}

	path, handler, err := instancesserver.NewHandler(v, servers.WithContextProvider(ctxProvider))
	if err != nil {
		return fmt.Errorf("unable to initialize installs server: %w", err)
	}

	mux.Handle(path, handler)
	return nil
}

// registerDeploymentsServer registers the installs service handler on the provided mux
func registerDeploymentsServer(v *validator.Validate, mux *http.ServeMux, cfg *internal.Config) error {
	ctxProvider, err := orgcontext.NewStaticProvider(v, orgcontext.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create orgcontext provider: %w", err)
	}

	path, handler, err := deploymentsserver.NewHandler(v, servers.WithContextProvider(ctxProvider))
	if err != nil {
		return fmt.Errorf("unable to initialize installs server: %w", err)
	}

	mux.Handle(path, handler)
	return nil
}

// registerInstallsServer registers the installs service handler on the provided mux
func registerInstallsServer(v *validator.Validate, mux *http.ServeMux, cfg *internal.Config) error {
	ctxProvider, err := orgcontext.NewStaticProvider(v, orgcontext.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create orgcontext provider: %w", err)
	}

	path, handler, err := installserver.NewHandler(v, servers.WithContextProvider(ctxProvider))
	if err != nil {
		return fmt.Errorf("unable to initialize installs server: %w", err)
	}

	mux.Handle(path, handler)
	return nil
}

// registerOrgsServer registers the orgs service handler on the provided mux
func registerOrgsServer(v *validator.Validate, mux *http.ServeMux, cfg *internal.Config) error {
	ctxProvider, err := orgcontext.NewStaticProvider(v, orgcontext.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create orgcontext provider: %w", err)
	}

	path, handler, err := orgsserver.NewHandler(v, servers.WithContextProvider(ctxProvider))
	if err != nil {
		return fmt.Errorf("unable to initialize status server: %w", err)
	}

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

	v := validator.New()
	mux := http.NewServeMux()
	if err := registerStatusServer(mux, &cfg); err != nil {
		l.Fatal("unable to register status server:", zap.Error(err))
	}
	if err := registerOrgsServer(v, mux, &cfg); err != nil {
		l.Fatal("unable to register orgs server:", zap.Error(err))
	}
	if err := registerAppsServer(v, mux, &cfg); err != nil {
		l.Fatal("unable to register apps server:", zap.Error(err))
	}
	if err := registerInstallsServer(v, mux, &cfg); err != nil {
		l.Fatal("unable to register installs server:", zap.Error(err))
	}
	if err := registerDeploymentsServer(v, mux, &cfg); err != nil {
		l.Fatal("unable to register installs server:", zap.Error(err))
	}
	if err := registerInstancesServer(v, mux, &cfg); err != nil {
		l.Fatal("unable to register installs server:", zap.Error(err))
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
