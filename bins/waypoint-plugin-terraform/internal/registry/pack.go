package registry

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"oras.land/oras-go/v2"
)

func (b *Registry) packDirectory(ctx context.Context, log hclog.Logger, status terminal.Status, filePaths []fileRef) error {
	fileDescriptors := make([]v1.Descriptor, 0, len(filePaths))
	for idx, f := range filePaths {
		fi, err := os.Stat(f.absPath)
		if err != nil {
			return fmt.Errorf("failed to stat %s: %w", f.absPath, err)
		}
		if fi.Size() == 0 {
			status.Step(terminal.StatusOK, fmt.Sprintf("%d skipping packing %s as %s - empty", idx, f.absPath, f.relPath))
			continue
		}

		status.Step(terminal.StatusOK, fmt.Sprintf("%d packing %s as %s", idx, f.absPath, f.relPath))
		fileDescriptor, err := b.Store.Add(ctx, f.relPath, defaultFileType, f.absPath)
		if err != nil {
			return fmt.Errorf("unable to pack %s: %w", f.absPath, err)
		}

		fileDescriptors = append(fileDescriptors, fileDescriptor)
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
	status.Step(terminal.StatusOK, fmt.Sprintf("tagged %s", defaultTag))
	log.Info("tagged %s", defaultTag)

	desc, err := b.Store.Resolve(ctx, defaultTag)
	if err != nil {
		return fmt.Errorf("unable to resolve tag: %w", err)
	}
	log.Info("found tag %s %v", defaultTag, desc)

	return nil
}
