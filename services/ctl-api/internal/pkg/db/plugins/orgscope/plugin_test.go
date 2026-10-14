package orgscope

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	callerOrg = "org_caller"
	otherOrg  = "org_other"
)

type Org struct {
	ID string `gorm:"primaryKey"`
}

type record struct {
	ID    string `gorm:"primaryKey"`
	Name  string
	OrgID string
	Org   Org
}

type bareOrgID struct {
	ID    string `gorm:"primaryKey"`
	OrgID string
}

func TestOrgScope(t *testing.T) {
	db := dryDB(t)
	caller := cctx.SetOrgIDContext(context.Background(), callerOrg)

	tests := map[string]struct {
		ctx        context.Context
		query      func(db *gorm.DB) *gorm.DB
		wantOrgSQL int
		wantCaller int
		wantWhere  bool
		wantWrapOR bool
	}{
		"id only": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("id = ?", "rec_1").Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"first by id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.First(&row, "id = ?", "rec_1")
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"where org_id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ?", callerOrg).Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"where org_id and id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ? AND id = ?", callerOrg, "rec_1").Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"where org_id and name": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ? AND name = ?", callerOrg, "acme").Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"where org_id and install_id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ? AND install_id = ?", callerOrg, "inst_1").Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"where org_id and id in": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ? AND id IN ?", callerOrg, []string{"rec_1", "rec_2"}).Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"where org_id and deleted_at": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ? AND deleted_at = 0", callerOrg).Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"chained org_id then id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ?", callerOrg).Where("id = ?", "rec_1").Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"org_id and name or id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ?", callerOrg).
					Where(db.Where("name = ?", "acme").Or("id = ?", "rec_1")).
					Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"raw org_id and id or name": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ? AND (id = ? or name = ?)", callerOrg, "rec_1", "acme").Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"struct org and id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where(record{OrgID: callerOrg, ID: "rec_1"}).Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"struct pointer org": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where(&record{OrgID: callerOrg}).Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"map org_id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where(map[string]any{"org_id": callerOrg, "id": "rec_1"}).Find(&row)
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"org_id or id still adds caller": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ? OR id = ?", callerOrg, "rec_1").Find(&row)
			},
			wantOrgSQL: 2,
			wantCaller: 2,
			wantWhere:  true,
			wantWrapOR: true,
		},
		"other org still adds caller": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("org_id = ?", otherOrg).Find(&row)
			},
			wantOrgSQL: 2,
			wantCaller: 1,
			wantWhere:  true,
		},
		"not org_id still adds caller": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("NOT org_id = ?", callerOrg).Find(&row)
			},
			wantOrgSQL: 2,
			wantCaller: 2,
			wantWhere:  true,
		},
		"auth api": {
			ctx: Skip(caller),
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("id = ?", "rec_1").Find(&row)
			},
			wantWhere: true,
		},
		"no org in context": {
			ctx: context.Background(),
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("id = ?", "rec_1").Find(&row)
			},
			wantWhere: true,
		},
		"bare org_id column": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				var row bareOrgID
				return db.Where("id = ?", "rec_1").Find(&row)
			},
			wantWhere: true,
		},
		"update id only": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				return db.Model(&record{}).Where("id = ?", "rec_1").Update("name", "acme")
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"update org_id and id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				return db.Model(&record{}).Where("org_id = ? AND id = ?", callerOrg, "rec_1").Update("name", "acme")
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"delete id only": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("id = ?", "rec_1").Delete(&record{})
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"delete org_id": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				return db.Where("org_id = ?", callerOrg).Delete(&record{})
			},
			wantOrgSQL: 1,
			wantCaller: 1,
			wantWhere:  true,
		},
		"create": {
			ctx: caller,
			query: func(db *gorm.DB) *gorm.DB {
				return db.Create(&record{ID: "rec_1", Name: "acme"})
			},
			wantOrgSQL: 1,
		},
		"association context skips": {
			ctx: context.WithValue(caller, associationScopeKey{}, true),
			query: func(db *gorm.DB) *gorm.DB {
				var row record
				return db.Where("id = ?", "rec_1").Find(&row)
			},
			wantWhere: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tx := tc.query(db.WithContext(tc.ctx))
			require.NoError(t, tx.Error)
			sql := tx.Statement.SQL.String()
			require.Equal(t, tc.wantOrgSQL, countOrgID(sql), sql)
			require.Equal(t, tc.wantCaller, countString(tx.Statement.Vars, callerOrg), sql)
			if tc.wantWhere {
				require.Contains(t, sql, "WHERE", sql)
			} else {
				require.NotContains(t, sql, "WHERE", sql)
			}
			if tc.wantWrapOR {
				require.Regexp(t, `(?i)\(.*\bor\b.*\).*\borg_id\b`, sql)
			}
		})
	}
}

func TestOrgScopeNextQueryStaysScoped(t *testing.T) {
	db := dryDB(t)
	ctx := cctx.SetOrgIDContext(context.Background(), callerOrg)

	var first record
	tx := db.WithContext(ctx).Where("id = ?", "rec_1").Find(&first)
	require.NoError(t, tx.Error)
	require.Equal(t, 1, countOrgID(tx.Statement.SQL.String()))

	var second record
	tx = db.WithContext(ctx).Where("id = ?", "rec_2").Find(&second)
	require.NoError(t, tx.Error)
	require.Equal(t, 1, countOrgID(tx.Statement.SQL.String()), tx.Statement.SQL.String())
	require.Equal(t, 1, countString(tx.Statement.Vars, callerOrg))
}

func dryDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "host=127.0.0.1 user=postgres dbname=postgres sslmode=disable",
	}), &gorm.Config{
		DryRun:                 true,
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	require.NoError(t, err)
	require.NoError(t, db.Use(New()))
	return db
}

func countOrgID(sql string) int {
	return len(orgIDColumnPattern.FindAllString(sql, -1))
}

func countString(vars []any, want string) int {
	n := 0
	for _, v := range vars {
		s, ok := v.(string)
		if ok && s == want {
			n++
		}
	}
	return n
}
