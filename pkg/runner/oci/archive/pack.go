package ociarchive

import (
	"context"
	"fmt"
	"path/filepath"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
	orasfile "oras.land/oras-go/v2/content/file"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/op"
)

const (
	defaultArtifactType string = "artifact/nuon"
	defaultLocalTag     string = "latest"
	tarLayerMediaType   string = "application/vnd.nuon.archive.layer.v1.tar+gzip"
	tarLayerTitle       string = "."
	maxManifestLayers          = 800
)

type FileRef struct {
	AbsPath  string `mapstructure:"abs_path,omitempty"`
	RelPath  string `mapstructure:"rel_path,omitempty"`
	FileType string `mapstructure:"file_type,omitempty"`
}

func (r *archive) Pack(ctx context.Context, log *zap.Logger, filePaths []FileRef) (retErr error) {
	opCtx, end := op.Tool(ctx, "oci", "pack")
	ctx = opCtx
	defer func() { end(retErr) }()
	if l, err := pkgctx.Logger(ctx); err == nil && l != nil {
		log = l
	}

	tarPath := filepath.Join(r.tmpDir, "archive-v1.tar.gz")
	if err := writeTarLayer(tarPath, filePaths); err != nil {
		return fmt.Errorf("unable to create tar layer: %w", err)
	}

	layer, err := r.store.Add(ctx, tarLayerTitle, tarLayerMediaType, tarPath)
	if err != nil {
		return fmt.Errorf("unable to add tar layer: %w", err)
	}
	layer.Annotations[orasfile.AnnotationUnpack] = "true"

	descriptor, err := packManifest(ctx, r.store, filePaths, []v1.Descriptor{layer})
	if err != nil {
		return err
	}

	if err := r.store.Tag(ctx, descriptor, defaultLocalTag); err != nil {
		return fmt.Errorf("unable to tag manifest: %w", err)
	}

	_, err = r.store.Resolve(ctx, defaultLocalTag)
	if err != nil {
		return fmt.Errorf("unable to resolve tag: %w", err)
	}
	log.Info("found tag", zap.String("tag", defaultLocalTag))

	return nil
}

func packManifest(ctx context.Context, store oras.Target, files []FileRef, layers []v1.Descriptor) (v1.Descriptor, error) {
	if len(layers) > maxManifestLayers {
		return v1.Descriptor{}, fmt.Errorf("refusing to pack %d files into %d OCI layers: layer count exceeds limit of %d", len(files), len(layers), maxManifestLayers)
	}

	descriptor, err := oras.Pack(ctx, store, defaultArtifactType, layers, oras.PackOptions{
		PackImageManifest: true,
		ManifestAnnotations: map[string]string{
			v1.AnnotationCreated: "1970-01-01T00:00:00Z",
		},
	})
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("unable to pack manifest: %w", err)
	}
	return descriptor, nil
}
