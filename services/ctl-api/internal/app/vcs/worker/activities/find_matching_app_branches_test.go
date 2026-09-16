package activities

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestMatchingAppBranchIgnoresRunConfigDuringScan(t *testing.T) {
	_, err := schema.Parse(&MatchingAppBranch{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatalf("parse MatchingAppBranch schema: %v", err)
	}
}
