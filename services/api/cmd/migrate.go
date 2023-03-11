package cmd

import (
	"errors"
	"log"

	_ "github.com/lib/pq"
	"github.com/powertoolsdev/mono/pkg/common/config"
	"github.com/powertoolsdev/mono/services/api/internal"
	databaseclient "github.com/powertoolsdev/mono/services/api/internal/clients/database"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate up|down|status",
	Short: "migrate the database using goose",
	Run:   runMigrate,
}

func init() { //nolint: gochecknoinits
	flags := migrateCmd.Flags()
	flags.String("service_name", "api", "the name of the service")
	flags.String("service_owner", "core", "the team that owns this service")
	flags.String("build_version", "", "the version build signature")
	flags.Bool("dry_run", true, "run with --dry-run=false to execute migrations")

	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) {
	var cfg internal.Config
	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		log.Fatal("failed to load config", zap.Error(err))
	}

	l, err := initializeLogger(&cfg)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if len(args) != 1 {
		l.Fatal("please provide an arg up|down|status", zap.Error(errors.New("incorrect arguments provided")))
	}

	if err := migrate(&cfg, args); err != nil {
		log.Fatal("failed to run migration", zap.Error(err))
	}
}

func migrate(cfg *internal.Config, args []string) error {
	db, err := databaseclient.New(databaseclient.WithConfig(cfg))
	if err != nil {
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	if err := goose.Run(args[0], sqlDB, cfg.DBMigrationsPath); err != nil {
		return err
	}

	return nil
}
