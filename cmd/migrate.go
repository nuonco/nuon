package cmd

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/powertoolsdev/api/internal"
	databaseclient "github.com/powertoolsdev/api/internal/clients/database"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/go-common/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate the database using gorm auto migrate",
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

	if err := migrate(&cfg, l); err != nil {
		log.Fatal("failed to run migration", zap.Error(err))
	}
}

func migrate(cfg *internal.Config, l *zap.Logger) error {
	db, err := databaseclient.New(databaseclient.WithConfig(cfg))
	if err != nil {
		return fmt.Errorf("unable to create database client: %w", err)
	}

	objs := []interface{}{
		&models.App{},
		&models.AWSSettings{},
		&models.Component{},
		&models.Deployment{},
		&models.Domain{},
		&models.GCPSettings{},
		&models.GithubConfig{},
		&models.Install{},
		&models.Org{},
		&models.User{},
		&models.UserOrg{},
	}

	for idx, o := range objs {
		l.Info(fmt.Sprintf("executing migration %v", idx))
		err = db.AutoMigrate(o)
		if err != nil {
			return fmt.Errorf("unable to execute migration %v: %w", idx, err)
		}
	}

	return nil
}
