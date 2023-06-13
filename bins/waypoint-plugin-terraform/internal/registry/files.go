package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

const (
	defaultFileType     string = "file/terraform"
	defaultArtifactType string = "artifact/terraform"
	defaultTag          string = "latest"
)

func (b *Registry) getSourceFiles(ctx context.Context, root string) ([]string, error) {
	fps := make([]string, 0)

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		fps = append(fps, path)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to walk %s: %w", root, err)
	}

	return fps, nil
}

func (b *Registry) packDirectory(ctx context.Context, log hclog.Logger, filePaths []string) error {
	fileDescriptors := make([]v1.Descriptor, len(filePaths))
	for idx, fp := range filePaths {
		log.Info("packing %s", fp)
		fileDescriptor, err := b.Store.Add(ctx, filepath.Base(fp), defaultFileType, fp)
		if err != nil {
			return fmt.Errorf("unable to pack %s: %w", fp, err)
		}

		fileDescriptors[idx] = fileDescriptor
	}

	descriptor, err := oras.Pack(ctx, b.Store, defaultArtifactType, fileDescriptors, oras.PackOptions{
		PackImageManifest: true,
	})
	if err != nil {
		return fmt.Errorf("unable to pack: %w", err)
	}

	if err := b.Store.Tag(ctx, descriptor, defaultTag); err != nil {
		return fmt.Errorf("unable to tag manifest: %w", err)
	}
	log.Info("tagged %s", defaultTag)

	desc, err := b.Store.Resolve(ctx, defaultTag)
	if err != nil {
		return fmt.Errorf("unable to resolve tag: %w", err)
	}
	log.Info("found tag %s %v", defaultTag, desc)

	return nil
}
