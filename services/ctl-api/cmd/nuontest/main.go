package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"

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

	// If go test args were provided, run go test with INTEGRATION=true
	if len(os.Args) > 1 {
		runGoTest(os.Args[1:])
	}
}

func runGoTest(args []string) {
	goTest := append([]string{"test"}, args...)

	log.Printf("running: go %v", goTest)

	cmd := exec.Command("go", goTest...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "INTEGRATION=true")

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		fmt.Fprintf(os.Stderr, "failed to run go test: %v\n", err)
		os.Exit(1)
	}
}
