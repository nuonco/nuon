package cmd

import (
	"fmt"
	"log"
	"net/http"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/interceptors"
	"github.com/powertoolsdev/mono/services/api/internal"
	databaseclient "github.com/powertoolsdev/mono/services/api/internal/clients/database"
	githubclient "github.com/powertoolsdev/mono/services/api/internal/clients/github"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"
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

type server struct {
	//
	v            *validator.Validate
	cfg          *internal.Config
	db           *gorm.DB
	gh           *ghinstallation.AppsTransport
	log          *zap.Logger
	interceptors []connect.Interceptor
}

var srvs []string = []string{
	// shared handlers
	"shared.v1.StatusService",

	// local handlers
	"admin.v1.AdminService",
	"app.v1.AppService",
	"component.v1.ComponentService",
	"deployment.v1.DeploymentService",
	"github.v1.GithubService",
	"install.v1.InstallService",
	"org.v1.OrgService",
	"user.v1.UserService",
}

// registerPrimaryServers registers each domain's server
//
//nolint:all
func newServer(cfg *internal.Config) (*server, error) {
	v := validator.New()
	l, err := initializeLogger(cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logger: %w", err)
	}

	tClient, err := temporal.New(v,
		temporal.WithAddr(cfg.TemporalHost),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize temporal: %w", err)
	}

	srv := &server{
		v:   v,
		cfg: cfg,
		log: l,
		interceptors: []connect.Interceptor{
			interceptors.LoggerInterceptor(),
			interceptors.NewTemporalClientInterceptor(tClient),
			interceptors.MetricsInterceptor(),
		},
	}

	db, err := databaseclient.New(databaseclient.WithConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to create database client: %w", err)
	}
	srv.db = db

	ghTransport, err := githubclient.New(githubclient.WithConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to github client: %w", err)
	}
	srv.gh = ghTransport

	return srv, nil
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
func runServer(cmd *cobra.Command, _ []string) {
	var cfg internal.Config

	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		log.Fatalf("unable to load config: %s", err)
	}

	srv, err := newServer(&cfg)
	if err != nil {
		log.Fatalf("unable to load server: %s", err)
	}

	mux := http.NewServeMux()
	if err := srv.registerAll(mux); err != nil {
		log.Fatalf("unable to register servers: %s", err)
	}

	srv.log.Info("server starting: ",
		zap.String("host", cfg.HTTPAddress),
		zap.String("port", cfg.HTTPPort))

	if err := http.ListenAndServe(
		fmt.Sprintf("%s:%s", cfg.HTTPAddress, cfg.HTTPPort),

		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		srv.log.Fatal("error on listen and server", zap.Error(err))
	}
}
