package migrations

import (
	"github.com/powertoolsdev/mono/services/ctl-api/internal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Migrations struct {
	db  *gorm.DB
	l   *zap.Logger
	cfg *internal.Config
}

func New(db *gorm.DB,
	cfg *internal.Config,
	l *zap.Logger) *Migrations {
	return &Migrations{
		db:  db,
		l:   l,
		cfg: cfg,
	}
}
