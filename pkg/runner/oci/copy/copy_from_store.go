package ocicopy

import (
	"context"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.opentelemetry.io/otel/attribute"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/pkg/runner/op"
)

func (c *copier) CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (_ *ocispec.Descriptor, retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "copy_from_store")
	ctx = opCtx
	defer func() { end(retErr) }()

	dstRepo, err := oci.GetRepo(ctx, dstCfg)
	if err != nil {
		return nil, err
	}

	spans := new(sync.Map)

	opts := oras.DefaultCopyOptions
	opts.PreCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		_, end := op.Start(ctx, "oci", "push_layer",
			attribute.String("oci.digest", string(desc.Digest)),
			attribute.String("oci.media_type", desc.MediaType),
			attribute.Int64("oci.size_bytes", desc.Size),
		)
		spans.Store(desc.Digest, end)
		return nil
	}
	opts.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		if endFn, ok := spans.LoadAndDelete(desc.Digest); ok {
			endFn.(op.EndFunc)(nil)
		}
		return nil
	}

	res, err := oras.Copy(ctx, store, srcTag, dstRepo, dstTag, opts)
	spans.Range(func(_, v any) bool {
		v.(op.EndFunc)(err)
		return true
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}
