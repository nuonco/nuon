package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/powertoolsdev/api/internal"
	databaseclient "github.com/powertoolsdev/api/internal/clients/database"
	githubclient "github.com/powertoolsdev/api/internal/clients/github"
	temporalclient "github.com/powertoolsdev/api/internal/clients/temporal"
	appsserver "github.com/powertoolsdev/api/internal/servers/apps"
	componentsserver "github.com/powertoolsdev/api/internal/servers/components"
	deploymentsserver "github.com/powertoolsdev/api/internal/servers/deployments"
	githubserver "github.com/powertoolsdev/api/internal/servers/github"
	installsserver "github.com/powertoolsdev/api/internal/servers/installs"
	orgsserver "github.com/powertoolsdev/api/internal/servers/orgs"
	statusserver "github.com/powertoolsdev/api/internal/servers/status"
	usersserver "github.com/powertoolsdev/api/internal/servers/users"
	"github.com/powertoolsdev/api/internal/services"
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
func registerPrimaryServers(mux *http.ServeMux, cfg *internal.Config, log *zap.Logger) error {
	db, err := databaseclient.New(databaseclient.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create database client: %w", err)
	}

	tc, err := temporalclient.New(temporalclient.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create temporal client: %w", err)
	}

	ghTransport, err := githubclient.New(githubclient.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to github client: %w", err)
	}

	appSvc := services.NewAppService(db, tc, log)
	_, err = appsserver.New(appsserver.WithHTTPMux(mux), appsserver.WithService(appSvc))
	if err != nil {
		return fmt.Errorf("unable to initialize apps server: %w", err)
	}

	componentsSvc := services.NewComponentService(db, log)
	_, err = componentsserver.New(componentsserver.WithHTTPMux(mux), componentsserver.WithService(componentsSvc))
	if err != nil {
		return fmt.Errorf("unable to initialize components server: %w", err)
	}

	deploymentsSvc := services.NewDeploymentService(db, tc, ghTransport, log)
	_, err = deploymentsserver.New(deploymentsserver.WithHTTPMux(mux), deploymentsserver.WithService(deploymentsSvc))
	if err != nil {
		return fmt.Errorf("unable to initialize deployments server: %w", err)
	}

	githubAppID, err := strconv.ParseInt(cfg.GithubAppID, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse github app id: %w", err)
	}
	appstp, err := ghinstallation.NewAppsTransport(http.DefaultTransport, githubAppID, []byte(cfg.GithubAppKey))
	if err != nil {
		return fmt.Errorf("unable to parse github app id: %w", err)
	}
	githubSvc := services.NewGithubService(appstp, log)
	_, err = githubserver.New(githubserver.WithHTTPMux(mux), githubserver.WithService(githubSvc))
	if err != nil {
		return fmt.Errorf("unable to initialize github server: %w", err)
	}

	installSvc := services.NewInstallService(db, tc, log)
	_, err = installsserver.New(installsserver.WithHTTPMux(mux), installsserver.WithService(installSvc))
	if err != nil {
		return fmt.Errorf("unable to initialize installs server: %w", err)
	}

	orgSvc := services.NewOrgService(db, tc, log)
	_, err = orgsserver.New(orgsserver.WithHTTPMux(mux), orgsserver.WithService(orgSvc))
	if err != nil {
		return fmt.Errorf("unable to initialize orgs server: %w", err)
	}

	userSvc := services.NewUserService(db, log)
	_, err = usersserver.New(usersserver.WithHTTPMux(mux), usersserver.WithService(userSvc))
	if err != nil {
		return fmt.Errorf("unable to initialize orgs server: %w", err)
	}

	return nil
}

func initializeLogger(cfg *internal.Config) (*zap.Logger, error) {
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
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logger: %w", err)
	}
	zap.ReplaceGlobals(l)
	return l, nil
}

//nolint:all
func runServer(cmd *cobra.Command, args []string) {
	var cfg internal.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		log.Fatalf("unable to load config: %s", err)
	}
	l, err := initializeLogger(&cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	mux := http.NewServeMux()
	if err := registerStatusServer(mux, &cfg); err != nil {
		l.Fatal("unable to register status server:", zap.Error(err))
	}
	registerLoadbalancerHealthCheck(mux)
	if err := registerPrimaryServers(mux, &cfg, l); err != nil {
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
