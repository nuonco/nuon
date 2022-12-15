package repos

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jaswdr/faker"
	"github.com/powertoolsdev/api/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	_db       *gorm.DB
	fakerSeed string
	fkr       faker.Faker
)

func init() { //nolint: gochecknoinits // removing this function will break testing suite currently
	flag.StringVar(&fakerSeed, "faker_seed", "0", "the seed timestamp for faker")
}

func TestMain(m *testing.M) {
	flag.Parse()
	seed, err := strconv.ParseInt(fakerSeed, 10, 64)
	if err != nil {
		panic(err)
	}
	if seed == 0 {
		seed = time.Now().Unix()
	}

	fmt.Printf("Creating faker with seed: %d\n", seed)
	fkr = faker.NewWithSeed(rand.NewSource(seed))

	exitCode := m.Run()
	os.Exit(exitCode)
}

func cleanupHook(testName string) func() {
	type entity struct {
		table string
		key   interface{}
	}
	var entries []entity
	hookName := fmt.Sprintf("nuon:test:%s", testName)

	err := _db.Callback().Create().After("gorm:after_create").Register(hookName, func(db *gorm.DB) {
		stmt := db.Statement
		schm := stmt.Schema

		id, ok := stmt.Dest.(models.IDer)
		if !ok {
			return
		}
		entries = append(entries, entity{
			table: schm.Table,
			key:   id.GetID(),
		})
	})
	if err != nil {
		log.Fatalln(err)
	}

	return func() {
		defer func() {
			err := _db.Callback().Create().Remove(hookName)
			if err != nil {
				log.Fatalln(err)
			}
		}()

		_, inTransaction := _db.ConnPool.(*sql.Tx)
		tx := _db
		if !inTransaction {
			tx = _db.Begin()
		}

		// delete in reverse order
		for i := len(entries) - 1; i >= 0; i-- {
			entry := entries[i]
			tx.Table(entry.table).Where("id = ?", entry.key).Delete("")
		}

		if !inTransaction {
			tx.Commit()
		}
	}
}

func testDB(testName string) (*gorm.DB, func()) { //nolint:gocritic //unnamedResult suggestion for returns
	// "memoize"
	if _db != nil {
		fn := cleanupHook(testName)
		return _db, fn
	}

	gcfg := &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             10 * time.Millisecond,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	}

	pg := postgres.New(postgres.Config{DSN: "dbname=api host=localhost user=postgres"})

	gormDB, err := gorm.Open(pg, gcfg)
	if err != nil {
		panic(err)
	}

	_db = gormDB
	return testDB(testName)
}

type repoTestState struct {
	orgRepo orgRepo

	db         *gorm.DB
	ctxCloseFn func()
}

// repoTest represents a repo test that initializes a user, org, app and each service
type repoTest struct {
	fn      func(context.Context, repoTestState)
	timeout time.Duration
	desc    string
	skip    bool
}

// execRepoTest executes a single repo test
func execRepoTest(t *testing.T, test repoTest) {
	db, dbCloseFn := testDB(t.Name())
	defer dbCloseFn()

	ctx := context.Background()
	ctx, ctxCloseFn := context.WithCancel(ctx)
	if test.timeout > time.Duration(0) {
		ctx, ctxCloseFn = context.WithTimeout(ctx, test.timeout)
	}
	defer ctxCloseFn()

	state := repoTestState{
		db:         db,
		orgRepo:    NewOrgRepo(db),
		ctxCloseFn: ctxCloseFn,
	}

	test.fn(ctx, state)
}

func execRepoTests(t *testing.T, tests []repoTest) {
	for idx, test := range tests {
		if test.skip {
			t.Logf("skipping test %d - %s", idx, test.desc)
			continue
		}

		t.Logf("starting test %d - %s", idx, test.desc)
		execRepoTest(t, test)
		t.Logf("finished test %d - %s", idx, test.desc)
	}
}
