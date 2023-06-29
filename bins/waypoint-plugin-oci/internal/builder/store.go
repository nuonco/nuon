package builder

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	defaultArtifactType string = "artifact/nuon"
	defaultFileType     string = "file/helm"
	defaultLocalTag     string = "latest"
)

type fileRef struct {
	absPath string
	relPath string
}

func (r *Builder) packArchive(ctx context.Context, log hclog.Logger, filePaths []fileRef) error {
	fileDescriptors := make([]v1.Descriptor, len(filePaths))

	for idx, f := range filePaths {
		fileDescriptor, err := r.store.Add(ctx, f.relPath, defaultFileType, f.absPath)
		if err != nil {
			return fmt.Errorf("unable to pack %s: %w", f.absPath, err)
		}

		fileDescriptors[idx] = fileDescriptor
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

	desc, err := r.store.Resolve(ctx, defaultLocalTag)
	if err != nil {
		return fmt.Errorf("unable to resolve tag: %w", err)
	}
	log.Info("found tag %s %v", defaultLocalTag, desc)

	return nil
}

func (r *Builder) pushArchive(ctx context.Context, accessInfo *ociv1.AccessInfo) error {
	repo, err := remote.NewRepository(accessInfo.Image)
	if err != nil {
		return fmt.Errorf("unable to get repository: %w", err)
	}

	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(accessInfo.Auth.ServerAddress, auth.Credential{
			Username: accessInfo.Auth.Username,
			Password: accessInfo.Auth.Password,
		}),
	}

	// 3. Copy from the file store to the remote repository
	_, err = oras.Copy(ctx, r.store, defaultLocalTag, repo, accessInfo.Tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy image to registry: %w", err)
	}

	return nil
}
