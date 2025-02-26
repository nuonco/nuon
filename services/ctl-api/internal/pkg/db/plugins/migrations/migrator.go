package migrations

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins"
)

type Opts struct {
	DropViewSQLTmpl   string
	CreateViewSQLTmpl string

	CreateIndexTmpl       string
	CreateUniqueIndexTmpl string
	CreatePKIndexTmpl     string
	DropIndexTmpl         string
}

func NewOpts() *Opts {
	return &Opts{
		DropViewSQLTmpl:   "DROP VIEW IF EXISTS %s",
		CreateViewSQLTmpl: "CREATE OR REPLACE VIEW %s AS %s",
		DropIndexTmpl:     "DROP INDEX IF EXISTS ?",

		CreateUniqueIndexTmpl: "CREATE UNIQUE INDEX ? ON ? (?)",
		CreatePKIndexTmpl:     "CREATE PRIMARY KEY INDEX ? ON ? (?)",
		CreateIndexTmpl:       "CREATE INDEX ? ON ? (?)",
	}
}

type Params struct {
	// Models
	Models     []any
	Migrations []Migration

	// Migrations DB is what is expected to have the migrations type registered
	MigrationsDB *gorm.DB

	TableOpts map[string]string
	Opts      *Opts

	// DB can be any gorm compatible db
	DB     *gorm.DB
	DBType string

	L   *zap.Logger
	Cfg *internal.Config
	MW  metrics.Writer
}

func New(params Params) *Migrator {
	return &Migrator{
		globalMigrations: params.Migrations,
		models:           params.Models,
		db:               params.DB,
		dbType:           params.DBType,
		l:                params.L,
		cfg:              params.Cfg,
		mw:               params.MW,
		migrationDB:      params.MigrationsDB,
		tableOpts:        params.TableOpts,
		opts:             params.Opts,
	}
}

type Migrator struct {
	opts             *Opts
	models           []any
	globalMigrations []Migration
	migrationDB      *gorm.DB
	db               *gorm.DB
	dbType           string
	tableOpts        map[string]string
	cfg              *internal.Config
	mw               metrics.Writer
	l                *zap.Logger
	allowDestroy     bool
}

func (m *Migrator) log(obj any) *zap.Logger {
	name := plugins.TableName(m.db, obj)

	return m.l.With(zap.String("model", name))
}
