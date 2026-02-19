package blobstore

import "context"

type blobContextKey string

const (
	BlobWriteEnabledKey blobContextKey = "blob_write_enabled"
	BlobAutoLoadKey     blobContextKey = "blob_auto_load"
	BlobServiceKey      blobContextKey = "blob_service"
)

// WithBlobWriteEnabled controls whether blobs upload to S3 on save
func WithBlobWriteEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, BlobWriteEnabledKey, enabled)
}

// WithBlobAutoLoad controls whether blobs auto-load from S3 on query
func WithBlobAutoLoad(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, BlobAutoLoadKey, enabled)
}

// WithBlobService sets the blobstore service in context
func WithBlobService(ctx context.Context, svc Service) context.Context {
	return context.WithValue(ctx, BlobServiceKey, svc)
}

// IsBlobWriteEnabled checks if blob writes are enabled (default: true)
func IsBlobWriteEnabled(ctx context.Context) bool {
	if v := ctx.Value(BlobWriteEnabledKey); v != nil {
		return v.(bool)
	}
	return true // Default: enabled
}

// IsBlobAutoLoad checks if auto-load is enabled (default: false)
func IsBlobAutoLoad(ctx context.Context) bool {
	if v := ctx.Value(BlobAutoLoadKey); v != nil {
		return v.(bool)
	}
	return false // Default: disabled (must be explicitly enabled)
}

// GetBlobService retrieves the blobstore service from context
func GetBlobService(ctx context.Context) Service {
	if v := ctx.Value(BlobServiceKey); v != nil {
		return v.(Service)
	}
	return nil
}
