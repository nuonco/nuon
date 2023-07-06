package builder

import (
	"fmt"
	"strings"

	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

func getRepo(image, username, password string) (*remote.Repository, error) {
	repo, err := remote.NewRepository(image)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

	registryHost := strings.SplitN(image, "/", 2)[0]
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(registryHost, auth.Credential{
			Username: username,
			Password: password,
		}),
	}

	return repo, nil
}
