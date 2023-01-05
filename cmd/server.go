package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/powertoolsdev/api/internal"
	appsserver "github.com/powertoolsdev/api/internal/servers/apps"
	componentsserver "github.com/powertoolsdev/api/internal/servers/components"
	deploymentsserver "github.com/powertoolsdev/api/internal/servers/deployments"
	installsserver "github.com/powertoolsdev/api/internal/servers/installs"
	orgsserver "github.com/powertoolsdev/api/internal/servers/orgs"
	statusserver "github.com/powertoolsdev/api/internal/servers/status"
	usersserver "github.com/powertoolsdev/api/internal/servers/users"
	"github.com/powertoolsdev/go-common/config"
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
	_, err := statusserver.New(statusserver.WithConfig(cfg), statusserver.WithHTTPMux(mux))
	if err != nil {
		return fmt.Errorf("unable to initialize status server: %w", err)
	}

	return nil
}

// registerPrimaryServers registers each domain's server
func registerPrimaryServers(mux *http.ServeMux, _ *internal.Config) error {
	_, err := appsserver.New(appsserver.WithHTTPMux(mux))
	if err != nil {
		return fmt.Errorf("unable to initialize apps server: %w", err)
	}

	_, err = componentsserver.New(componentsserver.WithHTTPMux(mux))
	if err != nil {
		return fmt.Errorf("unable to initialize components server: %w", err)
	}

	_, err = deploymentsserver.New(deploymentsserver.WithHTTPMux(mux))
	if err != nil {
		return fmt.Errorf("unable to initialize deployments server: %w", err)
	}

	_, err = installsserver.New(installsserver.WithHTTPMux(mux))
	if err != nil {
		return fmt.Errorf("unable to initialize installs server: %w", err)
	}

	_, err = orgsserver.New(orgsserver.WithHTTPMux(mux))
	if err != nil {
		return fmt.Errorf("unable to initialize orgs server: %w", err)
	}

	_, err = usersserver.New(usersserver.WithHTTPMux(mux))
	if err != nil {
		return fmt.Errorf("unable to initialize orgs server: %w", err)
	}

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
	if err := registerPrimaryServers(mux, &cfg); err != nil {
		l.Fatal("unable to register primary server:", zap.Error(err))
	}

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
