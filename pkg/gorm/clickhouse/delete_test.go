package clickhouse_test

import (
	"testing"
	"time"

	"gorm.io/gorm/utils/tests"
)

func TestDelete(t *testing.T) {
	user := User{ID: 2, Name: "delete", FirstName: "zhang", LastName: "jinzhu", Age: 18, Active: true, Salary: 8.8888}

	if err := DB.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user, got error %v", err)
	}

	var result User
	if err := DB.Find(&result, user.ID).Error; err != nil {
		t.Fatalf("failed to query user, got error %v", err)
	}

	tests.AssertEqual(t, result, user)

	if err := DB.Delete(&result).Error; err != nil {
		t.Fatalf("failed to delete user, got error %v", err)
	}

	time.Sleep(500 * time.Millisecond)
	if err := DB.First(&result, user.ID).Error; err == nil {
		t.Fatalf("should raise ErrRecordNotFound, got error %v", err)
	}
}
