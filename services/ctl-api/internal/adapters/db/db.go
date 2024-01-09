package db

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

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

func (d *database) Validate(v *validator.Validate) error {
	if err := v.Struct(d); err != nil {
		return fmt.Errorf("unable to validate database: %w", err)
	}

	return nil
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

func New(v *validator.Validate, l *zap.Logger, cfg *internal.Config) (*gorm.DB, error) {
	database := &database{
		Logger:     zapgorm2.New(l),
		PasswordFn: FetchIamTokenPassword,
		Host:       cfg.DBHost,
		User:       cfg.DBUser,
		Name:       cfg.DBName,
		Port:       cfg.DBPort,
		SSLMode:    cfg.DBSSLMode,
		Region:     cfg.DBRegion,
	}
	if cfg.DBPassword != "" {
		database.PasswordFn = nil
		database.Password = cfg.DBPassword
	}
	if err := database.Validate(v); err != nil {
		return nil, err
	}

	connCfg, err := database.connCfg()
	if err != nil {
		return nil, fmt.Errorf("unable to create database connection config: %w", err)
	}
	l.Info("conn config", zap.Any("cfg", connCfg.ConnString()))

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
		Logger:         database.Logger,
		TranslateError: true,
	}
	postgresCfg := postgres.Config{
		Conn: stdlib.OpenDB(*connCfg, stdlib.OptionBeforeConnect(beforeConnectFn)),
	}
	db, err := gorm.Open(postgres.New(postgresCfg), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := database.registerPlugins(db); err != nil {
		return nil, fmt.Errorf("unable to register plugins: %w", err)
	}

	return db, err
}
