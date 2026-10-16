package registryauth

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"go.uber.org/zap"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/google/externalaccount"
	"google.golang.org/api/impersonate"

	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
)

// azureTokenExchangeAudience is the audience the Azure managed identity token is
// minted for. It must match the `allowed_audiences` on the GCP Workload
// Identity provider created by the `nuonco/gar-access/google` Terraform module.
const azureTokenExchangeAudience = "api://AzureADTokenExchange"

const garUsername = "oauth2accesstoken"

var garScopes = []string{"https://www.googleapis.com/auth/cloud-platform"}

func (i *Issuer) GAR(ctx context.Context, l *zap.Logger, serviceAccountEmail, workloadIdentityProvider string) (*Token, error) {
	if workloadIdentityProvider != "" {
		if i.cfg.IsAzure() {
			return i.garViaAzureFederation(ctx, l, serviceAccountEmail, workloadIdentityProvider)
		}
		return i.garViaAWSFederation(ctx, serviceAccountEmail, workloadIdentityProvider)
	}

	if serviceAccountEmail != "" {
		ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: serviceAccountEmail,
			Scopes:          garScopes,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to impersonate %s: %w", serviceAccountEmail, err)
		}
		token, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("unable to get token for %s: %w", serviceAccountEmail, err)
		}
		return &Token{Username: garUsername, Password: token.AccessToken}, nil
	}

	creds, err := google.FindDefaultCredentials(ctx, garScopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to find GCP credentials: %w", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP access token: %w", err)
	}
	return &Token{Username: garUsername, Password: token.AccessToken}, nil
}

// awsCredentialSupplier implements externalaccount.AwsSecurityCredentialsSupplier
// using the AWS SDK's default credential chain (handles IRSA, IMDS, env vars).
type awsCredentialSupplier struct {
	region string
}

func (s *awsCredentialSupplier) AwsRegion(ctx context.Context, opts externalaccount.SupplierOptions) (string, error) {
	return s.region, nil
}

func (s *awsCredentialSupplier) AwsSecurityCredentials(ctx context.Context, opts externalaccount.SupplierOptions) (*externalaccount.AwsSecurityCredentials, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(s.region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}
	creds, err := cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get AWS credentials: %w", err)
	}
	return &externalaccount.AwsSecurityCredentials{
		AccessKeyID:     creds.AccessKeyID,
		SecretAccessKey: creds.SecretAccessKey,
		SessionToken:    creds.SessionToken,
	}, nil
}

func (i *Issuer) garViaAWSFederation(ctx context.Context, serviceAccountEmail, workloadIdentityProvider string) (*Token, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}
	region := awsCfg.Region
	if region == "" {
		region = "us-east-1"
	}

	return garFederatedToken(ctx, externalaccount.Config{
		Audience:                       fmt.Sprintf("//iam.googleapis.com/%s", workloadIdentityProvider),
		SubjectTokenType:               "urn:ietf:params:aws:token-type:aws4_request",
		Scopes:                         garScopes,
		AwsSecurityCredentialsSupplier: &awsCredentialSupplier{region: region},
	}, serviceAccountEmail)
}

// azureSubjectTokenSupplier feeds the Azure managed-identity OIDC token to GCP's
// STS as the subject token to exchange for a GAR access token.
type azureSubjectTokenSupplier struct {
	cred  azcore.TokenCredential
	scope string
}

func (s *azureSubjectTokenSupplier) SubjectToken(ctx context.Context, _ externalaccount.SupplierOptions) (string, error) {
	tok, err := s.cred.GetToken(ctx, policy.TokenRequestOptions{Scopes: []string{s.scope}})
	if err != nil {
		return "", fmt.Errorf("unable to get azure token: %w", err)
	}
	return tok.Token, nil
}

func (i *Issuer) garViaAzureFederation(ctx context.Context, l *zap.Logger, serviceAccountEmail, workloadIdentityProvider string) (*Token, error) {
	cred, err := azurecredentials.Fetch(ctx, nil, l)
	if err != nil {
		return nil, fmt.Errorf("unable to get azure credentials: %w", err)
	}

	return garFederatedToken(ctx, externalaccount.Config{
		Audience:         fmt.Sprintf("//iam.googleapis.com/%s", workloadIdentityProvider),
		SubjectTokenType: "urn:ietf:params:oauth:token-type:jwt",
		Scopes:           garScopes,
		SubjectTokenSupplier: &azureSubjectTokenSupplier{
			cred:  cred,
			scope: azureTokenExchangeAudience + "/.default",
		},
	}, serviceAccountEmail)
}

func garFederatedToken(ctx context.Context, cfg externalaccount.Config, serviceAccountEmail string) (*Token, error) {
	if serviceAccountEmail != "" {
		cfg.ServiceAccountImpersonationURL = fmt.Sprintf(
			"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken",
			serviceAccountEmail,
		)
	}

	ts, err := externalaccount.NewTokenSource(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create federated token source: %w", err)
	}

	token, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get federated token: %w", err)
	}

	return &Token{Username: garUsername, Password: token.AccessToken}, nil
}
