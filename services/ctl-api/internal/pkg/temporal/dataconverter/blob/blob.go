package blob

import (
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/filecache"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

const encoding = "nuon/blob"

var _ converter.PayloadCodec = (*dataConverter)(nil)

type dataConverter struct {
	cfg           *internal.Config
	l             *zap.Logger
	db            *gorm.DB
	blobSvc       blobstore.Service
	mw            metrics.Writer
	cache         *filecache.FileCache
	encodeEnabled bool
}
