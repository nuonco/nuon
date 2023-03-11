package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/powertoolsdev/mono/services/api/internal"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

type databaseOption func(*database) error

// database represents the set of configuration options for creating a database connection. If UseIAM is set, we will
// automatically create a database token using the AWS RDS api.
type database struct {
	User    string `validate:"required"`
	Host    string `validate:"required"`
	Name    string `validate:"required"`
	Port    string `validate:"required"`
	SSLMode string `validate:"required"`

	// required for IAM auth
	PasswordFn func(context.Context, database) (string, error)
	Region     string `validate:"required"`

	// required for local auth
	Password string

	Logger zapgorm2.Logger `validate:"required"`
}

func (c *database) connCfg() (*pgx.ConnConfig, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Host,
		c.User,
		c.Password,
		c.Name,
		c.Port,
		c.SSLMode)

	connCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	return connCfg, nil
}

func New(opts ...databaseOption) (*gorm.DB, error) {
	logger, _ := zap.NewProduction(zap.WithCaller(false))
	database := &database{
		Logger:     zapgorm2.New(logger),
		PasswordFn: FetchIamTokenPassword,
	}

	for idx, opt := range opts {
		if err := opt(database); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	connCfg, err := database.connCfg()
	if err != nil {
		return nil, err
	}

	beforeConnectFn := func(ctx context.Context, connCfg *pgx.ConnConfig) error {
		if database.PasswordFn == nil {
			return nil
		}

		password, er := database.PasswordFn(ctx, *database)
		if er != nil {
			return err
		}

		c := connCfg.Config
		c.Password = password
		connCfg.Config = c
		connCfg.Password = password
		return nil
	}

	gormCfg := &gorm.Config{
		Logger: database.Logger,
	}

	postgresCfg := postgres.Config{
		Conn: stdlib.OpenDB(*connCfg, stdlib.OptionBeforeConnect(beforeConnectFn)),
	}
	db, err := gorm.Open(postgres.New(postgresCfg), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	return db, err
}

func WithConfig(cfg *internal.Config) databaseOption {
	return func(d *database) error {
		d.Host = cfg.DBHost
		d.User = cfg.DBUser
		d.Name = cfg.DBName
		d.Port = cfg.DBPort
		d.SSLMode = cfg.DBSSLMode
		d.Region = cfg.DBRegion

		// if password is set, we disable the password function and use it
		if cfg.DBPassword != "" {
			d.PasswordFn = nil
			d.Password = cfg.DBPassword
		}

		return nil
	}
}

func WithLogger(log *zap.Logger) databaseOption {
	return func(d *database) error {
		d.Logger = zapgorm2.New(log)
		return nil
	}
}
