package registry

import (
	"context"
	"fmt"
	"strings"

	terraformv1 "github.com/powertoolsdev/mono/pkg/types/plugins/terraform/v1"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	defaultLocalTag string = "latest"
)

func (r *Registry) pushArtifact(ctx context.Context, accessInfo *terraformv1.AccessInfo) error {
	repo, err := remote.NewRepository(r.config.Repository)
	if err != nil {
		return fmt.Errorf("unable to get repository: %w", err)
	}

	registryHost := strings.SplitN(r.config.Repository, "/", 2)[0]
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(registryHost, auth.Credential{
			Username: accessInfo.Auth.Username,
			Password: accessInfo.Auth.Password,
		}),
	}

	// 3. Copy from the file store to the remote repository
	_, err = oras.Copy(ctx, r.Store, defaultLocalTag, repo, r.config.Tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("unable to copy image to registry: %w", err)
	}

	return nil
}
