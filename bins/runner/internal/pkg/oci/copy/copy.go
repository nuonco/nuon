package ocicopy

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"oras.land/oras-go/v2"

	pkgctx "github.com/powertoolsdev/mono/bins/runner/internal/pkg/ctx"
	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/oci"
	"github.com/powertoolsdev/mono/pkg/plugins/configs"
)

const (
	defaultCopyConcurrency int = 10
)

func (c *copier) Copy(ctx context.Context, srcCfg *configs.OCIRegistryRepository, srcTag string, dstCfg *configs.OCIRegistryRepository, dstTag string) (*ocispec.Descriptor, error) {
	srcRepo, err := oci.GetRepo(ctx, srcCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get source repo")
	}

	dstRepo, err := oci.GetRepo(ctx, dstCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get destination repo")
	}

	l, err := pkgctx.Logger(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get logger")
	}

	var hit bool

	timers := new(sync.Map)

	fields := func(desc ocispec.Descriptor) []zap.Field {
		return []zap.Field{
			zap.String("digest", string(desc.Digest)),
			zap.Int64("size", desc.Size),
		}
	}

	cpo := oras.CopyGraphOptions{
		Concurrency: defaultCopyConcurrency,
		OnCopySkipped: func(ctx context.Context, desc ocispec.Descriptor) error {
			hit = true
			l.Info(fmt.Sprintf("not copying %s of size %s, already present in repo (digest %s)", mediaNoun(desc.MediaType), humanize.Bytes(uint64(desc.Size)), desc.Digest), fields(desc)...)
			return nil
		},
		PreCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
			hit = true
			if desc.Digest != "" {
				timers.Store(desc.Digest, time.Now())
			}
			l.Info(fmt.Sprintf("copying %s of size %s (digest %s)", mediaNoun(desc.MediaType), humanize.Bytes(uint64(desc.Size)), desc.Digest), fields(desc)...)

			return nil
		},
		PostCopy: func(ctx context.Context, desc ocispec.Descriptor) error {
			hit = true
			if ti, has := timers.Load(desc.Digest); has {
				t := ti.(time.Time)
				l.Info(fmt.Sprintf("finished copying %s of size %s in %s (digest %s)", mediaNoun(desc.MediaType), humanize.Bytes(uint64(desc.Size)), time.Since(t), desc.Digest), fields(desc)...)
			} else {
				l.Info(fmt.Sprintf("finished copying %s (digest %s)", mediaNoun(desc.MediaType), desc.Digest), fields(desc)...)
			}
			return nil
		},
	}

	res, err := oras.Copy(ctx, srcRepo, srcTag, dstRepo, dstTag,
		oras.CopyOptions{
			CopyGraphOptions: cpo,
		})
	if err != nil {
		return nil, err
	}

	if !hit {
		l.Info("nothing to copy, all image layers already present in repo")
	}

	return &res, nil
}

func mediaNoun(mediatype string) string {
	switch mediatype {
	case ocispec.MediaTypeDescriptor:
		return "content descriptor"
	case ocispec.MediaTypeLayoutHeader:
		return "oci layout header"
	case ocispec.MediaTypeImageIndex:
		return "image index"
	case ocispec.MediaTypeImageManifest,
		"application/vnd.docker.distribution.manifest.v1+json",
		"application/vnd.docker.distribution.manifest.v2+json",
		"application/vnd.docker.distribution.manifest.list.v2+json":
		return "image manifest"
	case ocispec.MediaTypeImageConfig,
		"application/vnd.docker.container.image.v1+json":
		return "image config"
	case ocispec.MediaTypeImageLayer,
		ocispec.MediaTypeImageLayerGzip,
		ocispec.MediaTypeImageLayerZstd,
		"application/vnd.docker.image.rootfs.diff.tar.gzip",
		"application/vnd.docker.image.rootfs.foreign.diff.tar.gzip":
		return "image layer"
	case "application/vnd.docker.plugin.v1+json":
		return "plugin config"
	default:
		return "object"
	}
}
