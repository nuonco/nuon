package cmd

import (
	"log"
	"reflect"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/powertoolsdev/go-common/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/powertoolsdev/api/internal/database"
	"github.com/powertoolsdev/api/internal/models"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "migrate the database from db/migrations files",
	Run:   mig,
}

func init() { //nolint: gochecknoinits
	flags := migrateCmd.Flags()
	flags.String("service_name", "api", "the name of the service")
	flags.String("service_owner", "core", "the team that owns this service")
	flags.String("build_version", "", "the version build signature")

	rootCmd.AddCommand(migrateCmd)
}

func mig(cmd *cobra.Command, args []string) {
	var cfg Config
	if err := config.LoadInto(cmd.Flags(), &cfg); err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	logger.Info("configuring database connection for migrations")
	dbConfigFns := []database.DBConfigFunc{
		database.WithDBName(cfg.DBName),
		database.WithHost(cfg.DBHost),
		database.WithPort(cfg.DBPort),
		database.WithPassword(cfg.DBPassword),
		database.WithSSLMode(cfg.DBSSLMode),
		database.WithUser(cfg.DBUser),
		database.WithUseZap(cfg.DBZapLog),
		database.WithRegion(cfg.DBRegion),
	}

	if cfg.DBUseIAM {
		dbConfigFns = append(dbConfigFns, database.WithPasswordFn(database.FetchIamTokenPassword))
	}
	db, err := database.New(dbConfigFns...)
	if err != nil {
		logger.Fatal("error setting up database connection", zap.Error(err))
	}

	objs := []interface{}{
		&models.App{},
		&models.AWSSettings{},
		&models.Component{},
		&models.Deployment{},
		&models.Domain{},
		&models.GCPSettings{},
		&models.Install{},
		&models.Org{},
		&models.User{},
		&models.UserOrg{},
	}

	for _, o := range objs {
		logger.Info("migrating model:", zap.Any("type", reflect.TypeOf(o)))
		err = db.AutoMigrate(o)
		if err != nil {
			log.Fatalln(err)
		}
	}
}
