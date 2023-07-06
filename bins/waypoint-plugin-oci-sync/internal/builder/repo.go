package builder

import (
	"fmt"
	"path/filepath"
	"strings"

	ociv1 "github.com/powertoolsdev/mono/pkg/types/plugins/oci/v1"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func (b *Builder) getSrcRepo() (*remote.Repository, error) {
	// NOTE: when the source is configured, we know the full URL and pass the full path in here
	repo, err := remote.NewRepository(b.config.Image)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

	// NOTE(jm): the ServerAddress returned from the auth server usually can not be used - this is because AWS will
	// return a "preferred" address (for them), but the way most authentication for OCI clients works is by actually
	// pattern matching the authentication url with the base url of the image. If the image doesn't start with the
	// ServerAddress, then this will fail.
	registryURL := strings.Split(b.config.Image, "/")[0]
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(registryURL, auth.Credential{
			Username: b.config.Auth.Username,
			Password: b.config.Auth.AuthToken,
		}),
	}

	return repo, nil
}

func (b *Builder) getDstRepo(accessInfo *ociv1.AccessInfo) (*remote.Repository, error) {
	// NOTE: we don't know the id of the repo when this is generated, because we just know the name of the ecr repo.
	// Thus, the repo name at this point is not enough, and we have to join it with the registry address that was
	// returned from the auth.
	//
	// this is a bit indeterministic, because we're inferring the actual image url from the auth returned + the repo
	// name.
	//
	// TODO(jm): we should probably improve the way this all works and pass in the explicit image url to the plugin
	// -- however, that would require tracking the customer install's account ID in the executors, which we
	// currently don't have access too -- though, maybe we could pull it in from the sandbox outputs?
	repoURL := filepath.Join(accessInfo.Auth.ServerAddress, accessInfo.Image)

	repo, err := remote.NewRepository(repoURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(accessInfo.Auth.ServerAddress, auth.Credential{
			Username: accessInfo.Auth.Username,
			Password: accessInfo.Auth.Password,
		}),
	}

	return repo, nil
}
