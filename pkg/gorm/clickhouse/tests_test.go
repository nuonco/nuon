package clickhouse_test

import (
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/powertoolsdev/mono/pkg/gorm/clickhouse"
	"gorm.io/gorm"
)

var DB *gorm.DB

const dbDSN = "clickhouse://ctl-api:ctl-api@localhost:9000/test?dial_timeout=10s&read_timeout=20s"

func init() {
	integration := os.Getenv("GORM_INTEGRATION")
	if integration == "" {
		return
	}

	var err error

	if DB, err = gorm.Open(clickhouse.Open(dbDSN), &gorm.Config{}); err != nil {
		log.Printf("failed to connect database, got error %v", err)
		os.Exit(1)
	}

	RunMigrations()

	if os.Getenv("DEBUG") == "true" {
		DB = DB.Debug()
	}
}

func RunMigrations() {
	allModels := []interface{}{&User{}, &Log{}}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	rand.Shuffle(len(allModels), func(i, j int) { allModels[i], allModels[j] = allModels[j], allModels[i] })

	if err := DB.Migrator().DropTable(allModels...); err != nil {
		log.Printf("Failed to drop table, got error %v\n", err)
		os.Exit(1)
	}

	if err := DB.AutoMigrate(allModels...); err != nil {
		log.Printf("Failed to auto migrate, but got error %v\n", err)
		os.Exit(1)
	}

	for _, m := range allModels {
		if !DB.Migrator().HasTable(m) {
			log.Printf("Failed to create table for %#v\n", m)
			os.Exit(1)
		}
	}
}
