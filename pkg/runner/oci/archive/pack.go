package ociarchive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"

	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/runner/op"
)

const (
	defaultArtifactType string = "artifact/nuon"
	defaultLocalTag     string = "latest"

	// ECR caps an image at 500 layers, and we push one layer per source file.
	maxPerFileLayers int = 256
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

	fileDescriptors, err := r.addLayers(ctx, log, filePaths)
	if err != nil {
		return err
	}

	descriptor, err := oras.Pack(ctx, r.store, defaultArtifactType, fileDescriptors, oras.PackOptions{
		PackImageManifest: true,
	})
	if err != nil {
		return fmt.Errorf("unable to pack: %w", err)
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

func (r *archive) addLayers(ctx context.Context, log *zap.Logger, filePaths []FileRef) ([]v1.Descriptor, error) {
	if len(filePaths) > maxPerFileLayers {
		log.Info("packing source tree into a single tarball layer",
			zap.Int("file_count", len(filePaths)),
			zap.Int("max_per_file_layers", maxPerFileLayers),
		)

		tarballPath := filepath.Join(r.tmpDir, tarballName)
		if err := writeTarGz(tarballPath, filePaths); err != nil {
			return nil, fmt.Errorf("unable to write tarball: %w", err)
		}

		desc, err := r.store.Add(ctx, tarballName, tarballMediaType, tarballPath)
		if err != nil {
			return nil, fmt.Errorf("unable to pack tarball: %w", err)
		}
		return []v1.Descriptor{desc}, nil
	}

	fileDescriptors := make([]v1.Descriptor, 0, len(filePaths))

	for _, f := range filePaths {
		stat, err := os.Stat(f.AbsPath)
		if err != nil {
			return nil, fmt.Errorf("unable to stat file: %w", err)
		}

		if stat.Size() < 1 {
			log.Info("skipping empty file", zap.String("path", f.RelPath))
			continue
		}

		fileDescriptor, err := r.store.Add(ctx, f.RelPath, f.FileType, f.AbsPath)
		if err != nil {
			return nil, fmt.Errorf("unable to pack %s: %w", f.AbsPath, err)
		}

		fileDescriptors = append(fileDescriptors, fileDescriptor)
		log.Debug("packed file", zap.String("path", f.RelPath), zap.String("abspath", f.AbsPath))
	}

	return fileDescriptors, nil
}
