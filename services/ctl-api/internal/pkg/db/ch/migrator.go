package ch

import (
	"fmt"
	"regexp"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	chmigrations "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/ch/migrations"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type ChParams struct {
	fx.In

	Migrations   *chmigrations.Migrations
	MigrationsDB *gorm.DB `name:"psql"`
	DB           *gorm.DB `name:"ch"`

	L             *zap.Logger
	Cfg           *internal.Config
	MetricsWriter metrics.Writer
}

func NewCHMigrator(p ChParams, lc fx.Lifecycle) *migrations.Migrator {
	opts := migrations.NewOpts()

	cluster := p.Cfg.ClickhouseDBClusterName
	if cluster != "" {
		opts.CreateViewSQLTmpl = fmt.Sprintf("CREATE OR REPLACE VIEW %%s ON CLUSTER %s AS %%s", cluster)
	}
	// Empty cluster (ClickHouse Cloud): keep the default template, strip
	// ON CLUSTER clauses + downgrade Replicated engines via SQLRewriter.
	opts.SQLRewriter = clickhouseSQLRewriter(cluster)

	tableOpts := map[string]string{}
	if cluster != "" {
		tableOpts["gorm:table_cluster_options"] = "on cluster " + cluster
	}

	return migrations.New(migrations.Params{
		Opts:         opts,
		Migrations:   p.Migrations.All(),
		Models:       AllModels(),
		MigrationsDB: p.MigrationsDB,
		DB:           p.DB,
		DBType:       "ch",
		L:            p.L,
		Cfg:          p.Cfg,
		MW:           p.MetricsWriter,
		TableOpts:    tableOpts,
	})
}

var (
	// Match ON CLUSTER <identifier> with surrounding whitespace. Case-insensitive.
	onClusterRe = regexp.MustCompile(`(?i)\s+ON\s+CLUSTER\s+[A-Za-z0-9_]+`)
	// Match ReplicatedMergeTree('...', '...') (with optional whitespace) and
	// rewrite to MergeTree() — Cloud handles replication implicitly.
	replicatedMergeTreeRe = regexp.MustCompile(`(?i)Replicated([A-Za-z]*MergeTree)\s*\(\s*'[^']*'\s*,\s*'[^']*'\s*\)`)
)

// clickhouseSQLRewriter returns a function applied to every raw SQL
// string the migrator runs against ClickHouse. When cluster is empty
// (ClickHouse Cloud), it strips ON CLUSTER clauses and downgrades
// Replicated engines so the schema applies on Cloud. When cluster is
// set to a different name than the one baked into migrations, it
// rewrites the cluster name. When cluster matches the migration text,
// the rewriter is a no-op.
func clickhouseSQLRewriter(cluster string) func(string) string {
	return func(sql string) string {
		if cluster == "" {
			sql = onClusterRe.ReplaceAllString(sql, "")
			sql = replicatedMergeTreeRe.ReplaceAllStringFunc(sql, func(match string) string {
				// Replace `Replicated<X>MergeTree(...)` with `<X>MergeTree()`.
				sub := replicatedMergeTreeRe.FindStringSubmatch(match)
				if len(sub) >= 2 {
					return sub[1] + "()"
				}
				return "MergeTree()"
			})
			return sql
		}
		// Non-empty cluster: replace any ON CLUSTER token with the configured name.
		return onClusterRe.ReplaceAllString(sql, " ON CLUSTER "+cluster)
	}
}
