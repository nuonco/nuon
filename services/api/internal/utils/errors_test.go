package utils

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestIsDuplicateKeyError(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{DSN: "dbname=api host=localhost user=api"}), &gorm.Config{})
	assert.NoError(t, err)
	assert.NotNil(t, gormDB)

	db, err := gormDB.DB()
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// set up random table
	table := uuid.New()
	dupID := uuid.New()
	sql := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s"( id uuid DEFAULT gen_random_uuid(), PRIMARY KEY (id))`, table)

	_, err = db.Exec(sql)
	assert.NoError(t, err)
	defer func() {
		_, er := db.Exec(fmt.Sprintf(`DROP TABLE "%s"`, table))
		assert.NoError(t, er)
	}()

	_, err = db.Exec(fmt.Sprintf(`INSERT INTO "%s"(id) VALUES ('%s')`, table, dupID))
	assert.NoError(t, err)

	// Duplicate key is
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO "%s"(id) VALUES ('%s')`, table, dupID))
	assert.Error(t, err)
	assert.True(t, IsDuplicateKeyError(err))

	// Invalid value is not
	_, err = db.Exec(fmt.Sprintf(`INSERT INTO "%s"(id) VALUES (123)`, table))
	assert.Error(t, err)
	assert.False(t, IsDuplicateKeyError(err))

	// Random error is not
	assert.False(t, IsDuplicateKeyError(errors.New("this is not a duplicate key error")))
}
