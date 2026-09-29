package clickhouse_test

import (
	"log"
	"os"
	"testing"
	"time"

	chTypes "github.com/nuonco/nuon/pkg/gorm/clickhouse/pkg/types"
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

	if err := DB.Table("logs").AutoMigrate(&Log{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	type LogNestedUpdate1 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), value LowCardinality(String), valence LowCardinality(String));"`
	}
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate1{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	type LogNestedUpdate2 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence LowCardinality(String));"`
	}
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate2{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if DB.Migrator().HasColumn("logs", "events.value") {
		t.Fatalf("logs's `events`.`value` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	type LogNestedUpdate3 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence DateTime64(9));"`
	}
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate3{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	type LogNestedUpdate4 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence DateTime64(9), attributes Map(LowCardinality(String), String));"`
	}
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate4{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

	time.Sleep(50 * time.Millisecond)

	type LogNestedUpdate5 struct {
		ID        uint64 `gorm:"primaryKey"`
		CreatedAt time.Time
		UpdatedAt time.Time
		Name      string
		TraceID   string
		Events    chTypes.Nested `gorm:"type:Nested(key LowCardinality(String), valence DateTime64(9), attributes Map(LowCardinality(String), DateTime64(9)));"`
	}
	if err := DB.Table("logs").AutoMigrate(&LogNestedUpdate5{}); err != nil {
		t.Fatalf("no error should happen when auto migrate, but got %v", err)
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should exists after auto migrate")
	}

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

	if !DB.Migrator().HasColumn("logs", "events.key") {
		t.Fatalf("logs's `events`.`key` column should still exist after auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "events.valence") {
		t.Fatalf("logs's `events`.`valence` column should still exist after auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "events.attributes") {
		t.Fatalf("logs's `events`.`attributes` column should still exist after auto migrate")
	}

	if !DB.Migrator().HasColumn("logs", "links.time") {
		t.Fatalf("logs's `links`.`time` column should exists after auto migrate")
	}
	if !DB.Migrator().HasColumn("logs", "links.attributes") {
		t.Fatalf("logs's `links`.`attributes` column should exists after auto migrate")
	}
}
