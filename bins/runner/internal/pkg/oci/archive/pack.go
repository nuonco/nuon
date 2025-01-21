package ociarchive

import (
	"context"
	"fmt"
	"os"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"
)

const (
	defaultArtifactType string = "artifact/nuon"
	defaultLocalTag     string = "latest"
)

type FileRef struct {
	AbsPath  string `mapstructure:"abs_path,omitempty"`
	RelPath  string `mapstructure:"rel_path,omitempty"`
	FileType string `mapstructure:"file_type,omitempty"`
}

func (r *archive) Pack(ctx context.Context, log *zap.Logger, filePaths []FileRef) error {
	fileDescriptors := make([]v1.Descriptor, 0, len(filePaths))

	for _, f := range filePaths {
		stat, err := os.Stat(f.AbsPath)
		if err != nil {
			return fmt.Errorf("unable to stat file: %w", err)
		}

		if stat.Size() < 1 {
			log.Info("skipping empty file", zap.String("path", f.RelPath))
			continue
		}

		fileDescriptor, err := r.store.Add(ctx, f.RelPath, f.FileType, f.AbsPath)
		if err != nil {
			return fmt.Errorf("unable to pack %s: %w", f.AbsPath, err)
		}

		fileDescriptors = append(fileDescriptors, fileDescriptor)
		log.Info("packed file", zap.String("path", f.RelPath), zap.String("abspath", f.AbsPath))
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
