package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
)

const (
	defaultFileType     string = "file/terraform"
	defaultArtifactType string = "artifact/terraform"
	defaultTag          string = "latest"
)

type fileRef struct {
	absPath string
	relPath string
}

func (b *Builder) packDirectory(ctx context.Context, log hclog.Logger, filePaths []fileRef) error {
	fileDescriptors := make([]v1.Descriptor, len(filePaths))
	for idx, f := range filePaths {
		log.Error("packing %s %s", f.absPath, f.relPath)
		fileDescriptor, err := b.Store.Add(ctx, f.relPath, defaultFileType, f.absPath)
		if err != nil {
			return fmt.Errorf("unable to pack %s: %w", f.absPath, err)
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
