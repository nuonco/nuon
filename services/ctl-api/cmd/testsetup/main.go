package main

import (
	"log"

	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func main() {
	log.Println("setting up test databases...")

	// Create and migrate PostgreSQL test database
	dbCfg, err := tests.LoadDBConfig()
	if err != nil {
		log.Fatalf("failed to load db config: %v", err)
	}
	log.Printf("creating postgresql database %s...", dbCfg.DBName)
	err = tests.CreateAndMigrateDatabase(dbCfg)
	if err != nil {
		log.Fatalf("failed to setup postgresql: %v", err)
	}
	log.Println("postgresql database ready")

	// Create and migrate ClickHouse test database
	chCfg, err := tests.LoadCHConfig()
	if err != nil {
		log.Fatalf("failed to load clickhouse config: %v", err)
	}
	log.Printf("creating clickhouse database %s...", chCfg.Name)
	err = tests.CreateAndMigrateCHDatabase(chCfg)
	if err != nil {
		log.Fatalf("failed to setup clickhouse: %v", err)
	}
	log.Println("clickhouse database ready")

	log.Println("test database setup complete")
}
