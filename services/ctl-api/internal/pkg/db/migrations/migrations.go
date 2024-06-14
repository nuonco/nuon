package migrations

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/authz"
)

type Migrations struct {
	db          *gorm.DB
	l           *zap.Logger
	cfg         *internal.Config
	authzClient *authz.Client
}

func New(db *gorm.DB,
	cfg *internal.Config,
	authzClient *authz.Client,
	l *zap.Logger,
) *Migrations {
	return &Migrations{
		db:          db,
		l:           l,
		cfg:         cfg,
		authzClient: authzClient,
	}
}
