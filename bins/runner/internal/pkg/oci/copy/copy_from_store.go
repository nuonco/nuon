package ocicopy

import (
	"context"
	"sync"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.opentelemetry.io/otel/attribute"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/oci"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/op"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func (c *copier) CopyFromStore(ctx context.Context, store oras.ReadOnlyTarget, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (_ *ocispec.Descriptor, retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "copy_from_store")
	ctx = opCtx
	defer func() { end(retErr) }()

	dstRepo, err := oci.GetRepo(ctx, dstCfg)
	if err != nil {
		return nil, err
	}

	// spans holds the op.EndFunc per layer digest so PostCopy can finalize
	// the child span PreCopy opened. Per-layer pushes run concurrently via
	// oras's worker pool, so a sync.Map is required.
	spans := new(sync.Map)

	opts := oras.DefaultCopyOptions
	opts.PreCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		// Open a child of the surrounding oci.copy_from_store span so per-
		// layer push duration is visible in build traces (mirrors the
		// oci.pull_layer instrumentation in archive.Unpack).
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
	// Finalize any layer spans whose PostCopy never fired (failure path
	// or layers cancelled mid-flight). Done before the error return so
	// no spans leak.
	spans.Range(func(_, v any) bool {
		v.(op.EndFunc)(err)
		return true
	})
	if err != nil {
		return nil, err
	}

	return &res, nil
}
