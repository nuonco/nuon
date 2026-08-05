// Package blobsvc makes the blobstore service available to model query hooks.
//
// Blob-backed columns carry an S3 pointer rather than the payload, so the hooks
// that hydrate them need a service to read with. Injecting it for every query
// keeps a call path that forgot to plumb it from silently producing empty values.
package blobsvc

import (
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

var _ gorm.Plugin = (*plugin)(nil)

func NewPlugin(svc blobstore.Service) *plugin {
	return &plugin{svc: svc}
}

type plugin struct {
	svc blobstore.Service
}

func (p *plugin) Name() string {
	return "blob-service"
}

func (p *plugin) Initialize(db *gorm.DB) error {
	return db.Callback().Query().Before("gorm:query").Register("blob_service", p.inject)
}

// inject leaves an explicitly provided service alone, so a caller can still
// override it.
func (p *plugin) inject(db *gorm.DB) {
	if p.svc == nil || db.Statement == nil || db.Statement.Context == nil {
		return
	}
	if blobstore.GetBlobService(db.Statement.Context) != nil {
		return
	}

	db.Statement.Context = blobstore.WithBlobService(db.Statement.Context, p.svc)
}
