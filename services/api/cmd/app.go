package cmd

import (
	"fmt"

	ghinstallation "github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/bufbuild/connect-go"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/temporal/client"
	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/api/interceptors"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal"
	databaseclient "github.com/powertoolsdev/mono/services/api/internal/clients/database"
	githubclient "github.com/powertoolsdev/mono/services/api/internal/clients/github"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

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

type app struct {
	v            *validator.Validate
	cfg          *internal.Config
	db           *gorm.DB
	gh           *ghinstallation.AppsTransport
	log          *zap.Logger
	interceptors []connect.Interceptor
	tc           temporal.Client
	wfc          wfc.Client
}

func newApp(flags *pflag.FlagSet) (*app, error) {
	var cfg internal.Config
	if err := config.LoadInto(flags, &cfg); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}

	v := validator.New()
	if err := v.Struct(cfg); err != nil {
		return nil, fmt.Errorf("unable to validate config: %w", err)
	}

	l, err := initializeLogger(&cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize logger: %w", err)
	}

	tClient, err := temporal.New(v,
		temporal.WithAddr(cfg.TemporalHost),
		temporal.WithLogger(l),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize temporal: %w", err)
	}

	wfc, err := wfc.NewClient(v, wfc.WithClient(tClient))
	if err != nil {
		return nil, fmt.Errorf("unable to initialize workflows client: %w", err)
	}

	srv := &app{
		v:   v,
		cfg: &cfg,
		log: l,
		tc:  tClient,
		wfc: wfc,
		interceptors: []connect.Interceptor{
			interceptors.LoggerInterceptor(),
			interceptors.NewTemporalClientInterceptor(tClient),
			interceptors.MetricsInterceptor(),
		},
	}

	db, err := databaseclient.New(databaseclient.WithConfig(&cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to create database client: %w", err)
	}
	srv.db = db

	ghTransport, err := githubclient.New(githubclient.WithConfig(&cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to github client: %w", err)
	}
	srv.gh = ghTransport

	return srv, nil
}
