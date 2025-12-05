package clickhouse_test

import (
	"log"
	"os"
	"testing"
	"time"

	chTypes "github.com/powertoolsdev/mono/pkg/gorm/clickhouse/pkg/types"
)

type Log struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	TraceID   string
}

func TestAutoMigrateNested(t *testing.T) {
	integration := os.Getenv("GORM_INTEGRATION")
	if integration == "" {
		t.Skip("GORM_INTEGRATION=true must be set in environment to run.")
		return
	}

	log.Printf("[TestAutoMigrateNested] Testing Simple Nested Column Migration")
	type Log struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), value LowCardinality(String));"`
	}

	if err := DB.Table("logs").AutoMigrate(&Log{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}

	if !DB.Migrator().HasTable("logs") {
		t.Fatalf("logs should exists")
	}

	if DB.Migrator().HasColumn("logs", "events") {
		t.Fatalf("logs's events column should exists after first auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "events.key") {
		t.Fatalf("logs's `events`.`key` column should exists after auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "events.value") {
		t.Fatalf("logs's `events`.`value` column should exists after auto migrate")
	}

	columnTypes, err := DB.Migrator().ColumnTypes("logs")
	if err != nil {
		t.Fatalf("failed to get column types, got error %v", err)
	}

	for _, columnType := range columnTypes {
		switch columnType.Name() {
		case "id":
			if columnType.DatabaseTypeName() != "UInt64" {
				t.Fatalf("column id primary key should be correct, name: %v, column: %#v", columnType.Name(), columnType)
			}
		case "trace_id":
			if columnType.DatabaseTypeName() != "String" {
				t.Fatalf("column trace id should be correct, name: %v, column: %#v", columnType.Name(), columnType)
			}
		case "name":
			if columnType.DatabaseTypeName() != "String" {
				t.Fatalf("column name should be correct, name: %v, column: %#v", columnType.Name(), columnType)
			}
		}
	}

	// try auto migrating the model again. this second call to AutoMigrate should be a NoOp.
	if err := DB.Table("logs").AutoMigrate(&Log{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Test Addition
	type LogNestedUpdate1 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), value LowCardinality(String), valence LowCardinality(String));"`
	}
	// try auto migrating the model again. this second call to AutoMigrate should be a NoOp.
	// NOTE(fd): dev: automigrate again: expecting an error w/ events but we need to see/find that col name
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate1{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	// Test Deletion
	type LogNestedUpdate2 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence LowCardinality(String));"`
	}
	// try auto migrating the model again. this second call to AutoMigrate should be a NoOp.
	// NOTE(fd): dev: automigrate again: expecting an error w/ events but we need to see/find that col name
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate2{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if DB.Migrator().HasColumn("logs", "events.value") {
		t.Fatalf("logs's `events`.`value` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	// Test Modification
	type LogNestedUpdate3 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence DateTime64(9));"`
	}
	// try auto migrating the model again. this second call to AutoMigrate should be a NoOp.
	// NOTE(fd): dev: automigrate again: expecting an error w/ events but we need to see/find that col name
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate3{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	// Test Modification: Complex Nested Type
	type LogNestedUpdate4 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence DateTime64(9), attributes Map(LowCardinality(String), String));"`
	}
	// try auto migrating the model again. this second call to AutoMigrate should be a NoOp.
	// NOTE(fd): dev: automigrate again: expecting an error w/ events but we need to see/find that col name
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate4{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	// Test Modification: Complex Nested Type to a more Complex Nested Time
	type LogNestedUpdate5 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence DateTime64(9), attributes Map(LowCardinality(String), DateTime64(9)));"`
	}
	// try auto migrating the model again. this second call to AutoMigrate should be a NoOp.
	// NOTE(fd): dev: automigrate again: expecting an error w/ events but we need to see/find that col name
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate5{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

	// Test Modification: Column Deletion
	type LogNestedUpdate6 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Eventos   chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), value DateTime64(9), );"`
		Links     chTypes.Nested `gorm:"type:Nested( time DateTime64(9), attributes Map(LowCardinality(String), DateTime64(9)));"`
	}
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate6{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}

	// ensure old fields still exist: no destructive actions in automigration
	if !DB.Migrator().HasColumn("logs", "events.key") {
		t.Fatalf("logs's `events`.`key` column should still exist after auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should still exist after auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "events.attributes") {
		t.Fatalf("logs's `events`.`attributes` column should still exist after auto migrate")
	}

	// ensure new fields do exist
	if !DB.Migrator().HasColumn("logs", "links.time") {
		t.Fatalf("logs's `links`.`time` column should exists after auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "links.attributes") {
		t.Fatalf("logs's `links`.`attributes` column should exists after auto migrate")
	}
}
