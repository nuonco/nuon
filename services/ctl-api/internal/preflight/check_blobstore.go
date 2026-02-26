package preflight

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	internal "github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
)

// probeKey is never written. A not-found response still proves the bucket
// resolves and the credentials carry read access, which is what we want to
// verify without mutating the bucket.
const probeKey = ".nuon-preflight-probe"

var blobstoreCheck = Check{
	Name:        "blobstore",
	Description: "blob storage bucket access",

	Fields: func(cfg *internal.Config) []Field {
		return []Field{
			{Name: "blob_storage_provider", Value: cfg.BlobStorageProvider, Required: true},
			{Name: "blob_storage_bucket", Value: cfg.BlobStorageBucket, Required: true},
			{Name: "blob_storage_region", Value: cfg.BlobStorageRegion, Required: true},
		}
	},

	Probe: func(ctx context.Context, cfg *internal.Config) (string, error) {
		svc, err := blobstore.NewService(cfg, nopMetrics())
		if err != nil {
			return "", fmt.Errorf("unable to build blob storage client: %w", err)
		}

		_, _, err = svc.GetMetadata(ctx, probeKey)
		if err != nil && !isNotFound(err) {
			return "", fmt.Errorf("bucket unreachable: %w", err)
		}

		return fmt.Sprintf("bucket reachable %s", summary(
			"provider", cfg.BlobStorageProvider,
			"bucket", cfg.BlobStorageBucket,
			"region", cfg.BlobStorageRegion)), nil
	},
}

func isNotFound(err error) bool {
	var notFound *types.NotFound
	var noSuchKey *types.NoSuchKey

	return errors.As(err, &notFound) ||
		errors.As(err, &noSuchKey) ||
		errors.Is(err, storage.ErrObjectNotExist)
}
