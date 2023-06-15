package oci

import (
	"context"
	"fmt"

	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	defaultArtifactType string = "archive/terraform"
)

func (o *oci) getSrc() (oras.ReadOnlyTarget, error) {
	if o.testSrc != nil {
		return o.testSrc, nil
	}

	fmt.Println(o.Image.RepoURL())
	repo, err := remote.NewRepository(o.Image.RepoURL())
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(o.Image.Registry, auth.Credential{
			Username: o.Auth.Username,
			Password: o.Auth.Token,
		}),
	}

	return repo, nil
}

func (o *oci) pull(ctx context.Context) error {
	src, err := o.getSrc()
	if err != nil {
		return fmt.Errorf("unable to get repo client: %w", err)
	}

	manifest, err := oras.Copy(ctx, src, o.Image.Tag, o.store, o.Image.Tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy image: %w", err)
	}

	_, err = content.FetchAll(ctx, o.store, manifest)
	if err != nil {
		return fmt.Errorf("unable to fetch content: %w", err)
	}

	return nil
}
