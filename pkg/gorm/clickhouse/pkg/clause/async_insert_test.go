package clause_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	chClause "github.com/powertoolsdev/mono/pkg/gorm/clickhouse/pkg/clause"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID        uint64 `gorm:"primaryKey"`
	Name      string
	FirstName string
	LastName  string
	Age       int64 `gorm:"type:Nullable(Int64)"`
	Active    bool
	Salary    float32
	Attrs     map[string]string `gorm:"type:Map(String,String);"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func initDB() *gorm.DB {

	// borrowed from https://github.com/go-gorm/hints/blob/master/hints_test.go
	dummyDB, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DryRun: true,
	})

	chClause.Register(dummyDB)

	return dummyDB
}

func AssertSQL(t *testing.T, result *gorm.DB, sql string) {
	if result.Statement.SQL.String() != sql {
		fmt.Printf("statement: %s\n", result.Statement.SQL.String())
		fmt.Printf("result: %+v\n", result.Statement)
		t.Fatalf("SQL expects: %v, got %v", sql, result.Statement.SQL.String())
	}
}

func TestAsyncInsertClause(t *testing.T) {
	dummyDb := initDB()
	integration := os.Getenv("GORM_INTEGRATION")
	if integration == "" {
		t.Skip("GORM_INTEGRATION=true must be set in environment to run.")
		return
	}

	// sanity check
	result := dummyDb.Find(&User{})
	AssertSQL(t, result, "SELECT * FROM `users`")

	// test clause
	user := User{ID: 200, Name: "create", FirstName: "jerry", LastName: "smith", Age: 67, Active: true, Salary: 8.8888, Attrs: map[string]string{
		"a": "a",
		"b": "b",
	}}
	result = dummyDb.Clauses(chClause.AsyncInsert{}).Create(&user)
	AssertSQL(t, result, "INSERT INTO `users` SETTINGS async_insert=1, wait_for_async_insert=1 (`name`,`first_name`,`last_name`,`age`,`active`,`salary`,`attrs`,`created_at`,`updated_at`,`id`) VALUES (?,?,?,?,?,?,?,?,?,?) RETURNING `id`")

	// w/out async_insert
	user.ID = 201
	result = dummyDb.Create(&user)
	AssertSQL(t, result, "INSERT INTO `users` (`name`,`first_name`,`last_name`,`age`,`active`,`salary`,`attrs`,`created_at`,`updated_at`,`id`) VALUES (?,?,?,?,?,?,?,?,?,?) RETURNING `id`")
}
