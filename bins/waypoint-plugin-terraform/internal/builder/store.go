package builder

import (
	"context"
	"fmt"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
)

const (
	defaultStorePath    string = "/tmp"
	defaultFileType     string = "file/terraform"
	defaultArtifactType string = "artifact/terraform"
	defaultTag          string = "latest"
)

func (b *Builder) getStore() (*file.Store, error) {
	fs, err := file.New(defaultStorePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file store: %w", err)
	}

	return fs, nil
}

func (b *Builder) packDirectory(ctx context.Context, store *file.Store, filePaths []string) error {
	fileDescriptors := make([]v1.Descriptor, len(filePaths))
	for idx, fp := range filePaths {
		fileDescriptor, err := store.Add(ctx, fp, defaultFileType, "")
		if err != nil {
			return fmt.Errorf("unable to pack %s: %w", fp, err)
		}

		fileDescriptors[idx] = fileDescriptor
	}

	descriptor, err := oras.Pack(ctx, store, defaultArtifactType, fileDescriptors, oras.PackOptions{
		PackImageManifest: true,
	})
	if err != nil {
		return fmt.Errorf("unable to pack: %w", err)
	}

	if err := store.Tag(ctx, descriptor, defaultTag); err != nil {
		return fmt.Errorf("unable to tag manifest: %w", err)
	}

	return nil
}
