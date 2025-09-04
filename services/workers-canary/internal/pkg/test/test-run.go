package test

import (
	"fmt"

	"gorm.io/gorm"
)

// CanaryTestRun manages test execution with common properties handling
type CanaryTestRun struct {
	db *gorm.DB
	ct CanaryTests
}

func (e *CanaryTestRun) Execute() error {
	// TODO: create the canary group
	fmt.Println("Starting Canary Test Run for:", e.ct.GetProperties().Name)

	if err := e.ct.Setup(); err != nil {
		return err
	}
	results, err := e.ct.ExecTests()
	if err != nil {
		return err
	}
	// TODO: Handle the result (e.g., log it, and write to the database)
	fmt.Println("Canary Test Result:", results)

	if err := e.ct.Teardown(); err != nil {
		return err
	}
	return nil
}

func NewCanaryTestRun(db *gorm.DB, ct CanaryTests) *CanaryTestRun {
	return &CanaryTestRun{
		db: db,
		ct: ct,
	}
}
