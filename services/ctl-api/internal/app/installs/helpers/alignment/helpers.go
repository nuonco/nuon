package alignment

import (
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"

	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

type Helpers struct {
	db       *gorm.DB
	installs *installhelpers.Helpers
	blobSvc  blobstore.Service
	l        *zap.Logger
}

func New(db *gorm.DB, installs *installhelpers.Helpers, blobSvc blobstore.Service, l *zap.Logger) *Helpers {
	return &Helpers{db: db, installs: installs, blobSvc: blobSvc, l: l}
}

func (s *Helpers) WithDB(db *gorm.DB) *Helpers {
	scoped := *s
	scoped.db = db
	scoped.installs = s.installs.WithDB(db)
	return &scoped
}

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
