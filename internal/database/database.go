package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

type DBConfigFunc func(*Config)

// Config represents the set of configuration options for creating a database connection. If UseIAM is set, we will
// automatically create a database token using the AWS RDS api.
type Config struct {
	User       string
	Host       string
	Password   string
	DBName     string
	Port       string
	SSLMode    string
	UseZap     bool
	Region     string
	PasswordFn func(context.Context, Config) (string, error)
}

func (c *Config) String() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		c.Host,
		c.User,
		c.Password,
		c.DBName,
		c.Port,
		c.SSLMode)
}

func New(options ...DBConfigFunc) (*gorm.DB, error) {
	cfg := new(Config)
	cfg.UseZap = false

	var db *gorm.DB
	var err error

	for _, opt := range options {
		opt(cfg)
	}

	var dsn = cfg.String()
	connCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	beforeConnectFn := func(ctx context.Context, connCfg *pgx.ConnConfig) error {
		if cfg.PasswordFn == nil {
			return nil
		}

		password, er := cfg.PasswordFn(ctx, *cfg)
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
		Logger: logger.Default.LogMode(logger.Info),
	}
	if cfg.UseZap {
		logger := zapgorm2.New(zap.L())
		logger.SetAsDefault()
		gormCfg.Logger = logger
	}

	postgresCfg := postgres.Config{
		Conn: stdlib.OpenDB(*connCfg, stdlib.OptionBeforeConnect(beforeConnectFn)),
	}
	db, err = gorm.Open(postgres.New(postgresCfg), gormCfg)
	if err != nil {
		return nil, err
	}

	return db, err
}

func WithHost(s string) DBConfigFunc {
	return func(c *Config) {
		c.Host = s
	}
}

func WithSSLMode(s string) DBConfigFunc {
	return func(c *Config) {
		c.SSLMode = s
	}
}

func WithPort(s string) DBConfigFunc {
	return func(c *Config) {
		c.Port = s
	}
}

func WithUser(s string) DBConfigFunc {
	return func(c *Config) {
		c.User = s
	}
}

func WithDBName(s string) DBConfigFunc {
	return func(c *Config) {
		c.DBName = s
	}
}

func WithPassword(s string) DBConfigFunc {
	return func(c *Config) {
		c.Password = s
	}
}

func WithUseZap(b bool) DBConfigFunc {
	return func(c *Config) {
		c.UseZap = b
	}
}

func WithPasswordFn(fn func(context.Context, Config) (string, error)) DBConfigFunc {
	return func(c *Config) {
		c.PasswordFn = fn
	}
}

func WithRegion(s string) DBConfigFunc {
	return func(c *Config) {
		c.Region = s
	}
}
