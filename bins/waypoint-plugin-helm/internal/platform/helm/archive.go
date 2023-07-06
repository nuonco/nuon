package helm

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	ecrauthorization "github.com/powertoolsdev/mono/pkg/aws/ecr-authorization"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

const (
	defaultChartPackageFilename string = "chart.tgz"
)

func (o *Platform) initECRAuth(ctx context.Context) error {
	authProvider, err := ecrauthorization.New(o.v,
		ecrauthorization.WithCredentials(&credentials.Config{
			UseDefault: true,
		}),
		ecrauthorization.WithUseDefault(true),
	)
	if err != nil {
		return fmt.Errorf("unable to get auth provider: %w", err)
	}

	auth, err := authProvider.GetAuthorization(ctx)
	if err != nil {
		return fmt.Errorf("unable to get authorization: %w", err)
	}

	o.auth = auth
	return nil
}

func (o *Platform) getSrcRepo() (oras.ReadOnlyTarget, error) {
	baseURL := strings.TrimPrefix(o.auth.ServerAddress, "https://")
	repositoryURL := filepath.Join(baseURL, o.config.Archive.Image)
	repo, err := remote.NewRepository(repositoryURL)
	if err != nil {
		return nil, fmt.Errorf("unable to get repository: %w", err)
	}

	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(baseURL, auth.Credential{
			Username: o.auth.Username,
			Password: o.auth.RegistryToken,
		}),
	}

	return repo, nil
}

func (p *Platform) unpackArchive(ctx context.Context) (string, error) {
	if err := p.initECRAuth(ctx); err != nil {
		return "", fmt.Errorf("unable initialize ecr auth: %w", err)
	}
	p.log.Info("successfully initialized ECR auth")

	src, err := p.getSrcRepo()
	if err != nil {
		return "", fmt.Errorf("unable to get repo client: %w", err)
	}
	p.log.Info("successfully fetched source repo")

	manifest, err := oras.Copy(ctx, src, p.config.Archive.Tag, p.store, p.config.Archive.Tag, oras.DefaultCopyOptions)
	if err != nil {
		return "", fmt.Errorf("unable to copy image: %w", err)
	}
	p.log.Info("successfully copied manifest ")

	_, err = content.FetchAll(ctx, p.store, manifest)
	if err != nil {
		return "", fmt.Errorf("unable to fetch content: %w", err)
	}
	p.log.Info("successfully fetched manifest")

	return filepath.Join(p.tmpDir, "store", defaultChartPackageFilename), nil
}
