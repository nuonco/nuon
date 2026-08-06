package ociarchive

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	orasfile "oras.land/oras-go/v2/content/file"

	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/oci"
	"github.com/nuonco/nuon/pkg/runner/op"
)

func (a *archive) Unpack(ctx context.Context, srcCfg *configs.OCIRegistryRepository, tag string) (retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "unpack")
	ctx = opCtx
	defer func() { end(retErr) }()

	srcRepo, err := oci.GetRepo(ctx, srcCfg)
	if err != nil {
		return fmt.Errorf("unable to get source repo: %w", err)
	}
	return a.unpackFrom(ctx, srcRepo, tag)
}

// UnpackFromStore unpacks an archive artifact already present in a local
// store (e.g. an airgap bundle), bypassing all registry access.
func (a *archive) UnpackFromStore(ctx context.Context, src oras.ReadOnlyTarget, ref string) (retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "unpack_from_store")
	ctx = opCtx
	defer func() { end(retErr) }()

	return a.unpackFrom(ctx, src, ref)
}

func (a *archive) unpackFrom(ctx context.Context, src oras.ReadOnlyTarget, tag string) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return fmt.Errorf("unable to get logger: %w", err)
	}

	sourceManifest, err := src.Resolve(ctx, tag)
	if err != nil {
		return fmt.Errorf("unable to resolve source manifest: %w", err)
	}
	if err := validateArchiveManifest(ctx, src, sourceManifest); err != nil {
		return err
	}

	l.Info("pulling artifact from oci registry", zap.String("tag", tag))
	pullStart := time.Now()

	timers := new(sync.Map)
	// spans holds the op.EndFunc per layer digest so PostCopy can finalize
	// the child span PreCopy opened. Per-layer pulls run concurrently via
	// oras's worker pool, so a sync.Map is required.
	spans := new(sync.Map)
	var totalBytes int64

	fields := func(desc ocispec.Descriptor) []zap.Field {
		return []zap.Field{
			zap.String("digest", string(desc.Digest)),
			zap.String("media_type", desc.MediaType),
			zap.Int64("size", desc.Size),
		}
	}

	opts := oras.DefaultCopyOptions
	opts.PreCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		timers.Store(desc.Digest, time.Now())
		// Open a child of the surrounding oci.unpack span so per-layer
		// pull duration is visible in traces (e.g. to spot the bundled
		// terraform binaries layer dominating the pull).
		_, end := op.Start(ctx, "oci", "pull_layer",
			attribute.String("oci.digest", string(desc.Digest)),
			attribute.String("oci.media_type", desc.MediaType),
			attribute.Int64("oci.size_bytes", desc.Size),
		)
		spans.Store(desc.Digest, end)
		l.Info(
			fmt.Sprintf("pulling %s of size %s", desc.MediaType, humanize.Bytes(uint64(desc.Size))),
			fields(desc)...,
		)
		return nil
	}
	opts.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		totalBytes += desc.Size
		if endFn, ok := spans.LoadAndDelete(desc.Digest); ok {
			endFn.(op.EndFunc)(nil)
		}
		if ti, ok := timers.Load(desc.Digest); ok {
			t := ti.(time.Time)
			l.Info(
				fmt.Sprintf("finished pulling %s of size %s in %s",
					desc.MediaType, humanize.Bytes(uint64(desc.Size)), time.Since(t)),
				fields(desc)...,
			)
		}
		return nil
	}
	opts.OnCopySkipped = func(ctx context.Context, desc ocispec.Descriptor) error {
		l.Info(
			fmt.Sprintf("skipping %s of size %s, already present locally",
				desc.MediaType, humanize.Bytes(uint64(desc.Size))),
			fields(desc)...,
		)
		return nil
	}

	manifest, err := oras.Copy(ctx, src, sourceManifest.Digest.String(), a.store, tag, opts)
	// Finalize any layer spans whose PostCopy never fired (typically the
	// failure case below; PreCopy may also have started spans for layers
	// the caller cancelled mid-flight). Done before the error return so
	// no spans leak on the failure path.
	spans.Range(func(_, v any) bool {
		v.(op.EndFunc)(err)
		return true
	})
	if err != nil {
		return fmt.Errorf("unable to copy image: %w", err)
	}

	l.Info(
		fmt.Sprintf("finished pulling artifact (%s across all layers) in %s",
			humanize.Bytes(uint64(totalBytes)), time.Since(pullStart)),
		zap.Int64("total_bytes", totalBytes),
		zap.String("manifest_digest", string(manifest.Digest)),
	)

	fetchStart := time.Now()
	l.Info("fetching artifact contents into local store")
	if _, err = content.FetchAll(ctx, a.store, manifest); err != nil {
		return fmt.Errorf("unable to fetch contents: %w", err)
	}
	l.Info("finished fetching artifact contents", zap.String("duration", time.Since(fetchStart).String()))

	return nil
}

func validateArchiveManifest(ctx context.Context, fetcher content.Fetcher, descriptor ocispec.Descriptor) error {
	if descriptor.MediaType != ocispec.MediaTypeImageManifest {
		return fmt.Errorf("unsupported OCI archive manifest media type %q", descriptor.MediaType)
	}

	data, err := content.FetchAll(ctx, fetcher, descriptor)
	if err != nil {
		return fmt.Errorf("unable to fetch source manifest: %w", err)
	}
	var manifest ocispec.Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("unable to decode source manifest: %w", err)
	}
	if manifest.Config.MediaType != defaultArtifactType {
		return fmt.Errorf("unsupported OCI archive artifact type %q", manifest.Config.MediaType)
	}
	if len(manifest.Layers) != 1 {
		return fmt.Errorf("unsupported OCI archive layer count %d: expected 1", len(manifest.Layers))
	}

	layer := manifest.Layers[0]
	if layer.MediaType != tarLayerMediaType || layer.Annotations[ocispec.AnnotationTitle] != tarLayerTitle || layer.Annotations[orasfile.AnnotationUnpack] != "true" {
		return fmt.Errorf("unsupported OCI archive layer format %q", layer.MediaType)
	}
	return nil
}
