package cctx

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
)

var ErrBlobServiceContextNotFound error = fmt.Errorf("blob service context not found")

func BlobServiceFromGinContext(ctx *gin.Context) (blobstore.Service, error) {
	svc := ctx.Value(keys.BlobServiceCtxKey)
	if svc == nil {
		return nil, ErrBlobServiceContextNotFound
	}

	return svc.(blobstore.Service), nil
}

func SetBlobServiceGinContext(ctx *gin.Context, svc blobstore.Service) {
	ctx.Set(keys.BlobServiceCtxKey, svc)
}
