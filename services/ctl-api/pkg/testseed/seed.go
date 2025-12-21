package testseed

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
)

// Params defines the dependencies required by the Seeder.
// This follows the FX dependency injection pattern used throughout ctl-api.
type Params struct {
	fx.In

	L          *zap.Logger
	DB         *gorm.DB `name:"psql"`
	AcctClient *account.Client
}

// Seeder provides helper methods for seeding test data in integration tests.
// It manages the creation of test fixtures with proper context management.
type Seeder struct {
	db          *gorm.DB
	l           *zap.Logger
	acctHelpers *account.Client
}

// New creates a new Seeder instance with the given dependencies.
// This is called by FX during test setup.
func New(params Params) *Seeder {
	return &Seeder{
		db:          params.DB,
		l:           params.L,
		acctHelpers: params.AcctClient,
	}
}
