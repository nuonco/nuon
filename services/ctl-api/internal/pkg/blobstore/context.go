package blobstore

import (
	"context"

	"github.com/gin-gonic/gin"
)

type blobContextKey string

const (
	BlobWriteEnabledKey blobContextKey = "blob_write_enabled"
	BlobReadEnabledKey  blobContextKey = "blob_read_enabled"
	BlobAutoLoadKey     blobContextKey = "blob_auto_load"
	BlobServiceKey      blobContextKey = "blob_service"
)

func WithBlobWriteEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, BlobWriteEnabledKey, enabled)
}

func WithBlobReadEnabled(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, BlobReadEnabledKey, enabled)
}

func WithBlobAutoLoad(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, BlobAutoLoadKey, enabled)
}

func WithBlobService(ctx context.Context, svc Service) context.Context {
	return context.WithValue(ctx, BlobServiceKey, svc)
}

func WithBlobServiceGin(ctx *gin.Context, svc Service) context.Context {
	return context.WithValue(ctx, BlobServiceKey, svc)
}

func IsBlobWriteEnabled(ctx context.Context) bool {
	if v := ctx.Value(BlobWriteEnabledKey); v != nil {
		return v.(bool)
	}
	return true
}

func IsBlobReadEnabled(ctx context.Context) bool {
	if v := ctx.Value(BlobReadEnabledKey); v != nil {
		return v.(bool)
	}
	return false
}

func IsBlobAutoLoad(ctx context.Context) bool {
	if v := ctx.Value(BlobAutoLoadKey); v != nil {
		return v.(bool)
	}
	return false
}

func GetBlobService(ctx context.Context) Service {
	if v := ctx.Value(BlobServiceKey); v != nil {
		return v.(Service)
	}
	return nil
}
