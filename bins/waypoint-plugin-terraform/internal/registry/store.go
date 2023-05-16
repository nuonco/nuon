package registry

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	defaultStorePath string = "/tmp"
	defaultLocalTag  string = "latest"
)

func (r *Registry) getStore() (*file.Store, error) {
	fs, err := file.New(defaultStorePath)
	if err != nil {
		return nil, fmt.Errorf("unable to open file store: %w", err)
	}

	return fs, nil
}

func (r *Registry) pushArtifact(ctx context.Context, store *file.Store, accessInfo *AccessInfo) error {
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
	_, err = oras.Copy(ctx, store, defaultLocalTag, repo, r.config.Tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy image to registry: %w", err)
	}

	return nil
}
