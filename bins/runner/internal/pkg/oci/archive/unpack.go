package ociarchive

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"

	pkgctx "github.com/nuonco/nuon/bins/runner/internal/pkg/ctx"
	"github.com/nuonco/nuon/bins/runner/internal/pkg/oci"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func (a *archive) Unpack(ctx context.Context, srcCfg *configs.OCIRegistryRepository, tag string) error {
	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return fmt.Errorf("unable to get logger: %w", err)
	}

	srcRepo, err := oci.GetRepo(ctx, srcCfg)
	if err != nil {
		return fmt.Errorf("unable to get source repo: %w", err)
	}

	l.Info("pulling artifact from oci registry", zap.String("tag", tag))
	pullStart := time.Now()

	timers := new(sync.Map)
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
		l.Info(
			fmt.Sprintf("pulling %s of size %s", desc.MediaType, humanize.Bytes(uint64(desc.Size))),
			fields(desc)...,
		)
		return nil
	}
	opts.PostCopy = func(ctx context.Context, desc ocispec.Descriptor) error {
		totalBytes += desc.Size
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

	manifest, err := oras.Copy(ctx, srcRepo, tag, a.store, tag, opts)
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
