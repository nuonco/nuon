// Package orgregistry resolves and mints credentials for an app's repository in
// the management registry.
package orgregistry

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	ecrauthorization "github.com/nuonco/nuon/pkg/aws/ecr-authorization"
	"github.com/nuonco/nuon/pkg/aws/ecr/sessionpolicy"
	"github.com/nuonco/nuon/pkg/azure/acr"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	garUsername = "oauth2accesstoken"
	garScope    = "https://www.googleapis.com/auth/cloud-platform"
)

type Credential struct {
	Username string
	Password string
	// ServerAddress is scheme-less so it can prefix-match the repository URI.
	ServerAddress string
	Repository    string
}

type Client struct {
	v   *validator.Validate
	db  *gorm.DB
	cfg *internal.Config
}

func New(v *validator.Validate, db *gorm.DB, cfg *internal.Config) *Client {
	return &Client{v: v, db: db, cfg: cfg}
}

func (c *Client) Repository(ctx context.Context, appID string) (*configs.OCIRegistryRepository, error) {
	repo, err := c.appRepository(ctx, appID)
	if err != nil {
		return nil, err
	}

	loginServer, repository, err := splitRepositoryURI(repo.RepositoryURI)
	if err != nil {
		return nil, err
	}

	return &configs.OCIRegistryRepository{
		Plugin:            "oci",
		RegistryType:      configs.OCIRegistryTypePrivateOCI,
		LoginServer:       loginServer,
		Repository:        repository,
		ManagedCredential: true,
	}, nil
}

func (c *Client) Matches(ctx context.Context, appID string, cfg *configs.OCIRegistryRepository) (bool, error) {
	expected, err := c.Repository(ctx, appID)
	if err != nil {
		return false, err
	}
	if cfg == nil {
		return false, nil
	}
	return strings.TrimPrefix(cfg.LoginServer, "https://") == expected.LoginServer &&
		cfg.Repository == expected.Repository, nil
}

// GAR cannot scope a token below the registry, so a GCP control plane still
// hands out its own identity's token.
func (c *Client) Mint(ctx context.Context, orgID, appID string, push bool) (*Credential, error) {
	repo, err := c.appRepository(ctx, appID)
	if err != nil {
		return nil, err
	}
	if repo.OrgID != orgID {
		return nil, fmt.Errorf("app repository %s does not belong to org %s", repo.ID, orgID)
	}

	loginServer, repository, err := splitRepositoryURI(repo.RepositoryURI)
	if err != nil {
		return nil, err
	}

	var cred *Credential
	switch {
	case c.cfg.IsGCP():
		cred, err = c.mintGAR(ctx)
	case c.cfg.IsAzure():
		cred, err = c.mintACR(ctx, repository, push)
	default:
		cred, err = c.mintECR(ctx, repo, push)
	}
	if err != nil {
		return nil, err
	}

	cred.ServerAddress = loginServer
	cred.Repository = repository
	return cred, nil
}

func (c *Client) appRepository(ctx context.Context, appID string) (*app.AppRepository, error) {
	var repo app.AppRepository
	if err := c.db.WithContext(ctx).
		Where(app.AppRepository{AppID: appID}).
		First(&repo).Error; err != nil {
		return nil, fmt.Errorf("unable to get app repository for app %s: %w", appID, err)
	}
	return &repo, nil
}

func (c *Client) mintECR(ctx context.Context, repo *app.AppRepository, push bool) (*Credential, error) {
	repositoryARN, err := sessionpolicy.RepositoryARN(repo.RepositoryURI)
	if err != nil {
		return nil, fmt.Errorf("unable to build ecr repository arn: %w", err)
	}
	policy, err := sessionpolicy.Document(repositoryARN, push)
	if err != nil {
		return nil, err
	}

	ecr, err := ecrauthorization.New(c.v,
		ecrauthorization.WithCredentials(&credentials.Config{
			Region: repo.Region,
			AssumeRole: &credentials.AssumeRoleConfig{
				RoleARN:       c.cfg.ManagementIAMRoleARN,
				SessionName:   fmt.Sprintf("oci-sync-%s", repo.OrgID),
				SessionPolicy: policy,
			},
		}),
		ecrauthorization.WithRegistryID(c.cfg.ManagementECRRegistryID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create ecr authorizer: %w", err)
	}

	auth, err := ecr.GetAuthorization(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get ecr authorization: %w", err)
	}

	return &Credential{Username: auth.Username, Password: auth.RegistryToken}, nil
}

func (c *Client) mintGAR(ctx context.Context) (*Credential, error) {
	ts, err := google.DefaultTokenSource(ctx, garScope)
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP token source for GAR: %w", err)
	}
	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP access token for GAR: %w", err)
	}
	return &Credential{Username: garUsername, Password: token.AccessToken}, nil
}

func (c *Client) mintACR(ctx context.Context, repository string, push bool) (*Credential, error) {
	acrService := c.cfg.ManagementACRRegistryURL

	refreshToken, err := acr.GetRepositoryToken(ctx, nil, acrService, zap.L())
	if err != nil {
		return nil, fmt.Errorf("unable to get ACR repository token: %w", err)
	}

	actions := "pull"
	if push {
		actions = "pull,push"
	}
	token, err := acr.AccessToken(ctx, acrService, "repository:"+repository+":"+actions, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("unable to scope ACR access token: %w", err)
	}

	return &Credential{Username: acr.DefaultACRUsername, Password: token}, nil
}

func splitRepositoryURI(uri string) (loginServer, repository string, err error) {
	uri = strings.TrimPrefix(uri, "https://")
	loginServer, repository, ok := strings.Cut(uri, "/")
	if !ok || loginServer == "" || repository == "" {
		return "", "", fmt.Errorf("invalid app repository uri %q", uri)
	}
	return loginServer, repository, nil
}
