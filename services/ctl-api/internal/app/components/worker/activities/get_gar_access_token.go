package activities

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/google/externalaccount"
	"google.golang.org/api/impersonate"
)

type GetGARAccessTokenRequest struct {
	ServiceAccountEmail      string
	WorkloadIdentityProvider string
}

type GARAccessToken struct {
	Username string
	Password string
}

// @temporal-gen-v2 activity
// @max-retries 1
func (a *Activities) GetGARAccessToken(ctx context.Context, req *GetGARAccessTokenRequest) (*GARAccessToken, error) {
	scopes := []string{"https://www.googleapis.com/auth/cloud-platform"}

	if req.WorkloadIdentityProvider != "" {
		return a.getGARTokenViaFederation(ctx, req, scopes)
	}

	if req.ServiceAccountEmail != "" {
		ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: req.ServiceAccountEmail,
			Scopes:          scopes,
		})
		if err != nil {
			return nil, fmt.Errorf("unable to impersonate %s: %w", req.ServiceAccountEmail, err)
		}
		token, err := ts.Token()
		if err != nil {
			return nil, fmt.Errorf("unable to get token for %s: %w", req.ServiceAccountEmail, err)
		}
		return &GARAccessToken{Username: "oauth2accesstoken", Password: token.AccessToken}, nil
	}

	creds, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return nil, fmt.Errorf("unable to find GCP credentials: %w", err)
	}
	token, err := creds.TokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("unable to get GCP access token: %w", err)
	}
	return &GARAccessToken{Username: "oauth2accesstoken", Password: token.AccessToken}, nil
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

func (a *Activities) getGARTokenViaFederation(ctx context.Context, req *GetGARAccessTokenRequest, scopes []string) (*GARAccessToken, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}
	region := awsCfg.Region
	if region == "" {
		region = "us-east-1"
	}

	audience := fmt.Sprintf("//iam.googleapis.com/%s", req.WorkloadIdentityProvider)

	cfg := externalaccount.Config{
		Audience:                       audience,
		SubjectTokenType:               "urn:ietf:params:aws:token-type:aws4_request",
		Scopes:                         scopes,
		AwsSecurityCredentialsSupplier: &awsCredentialSupplier{region: region},
	}

	if req.ServiceAccountEmail != "" {
		cfg.ServiceAccountImpersonationURL = fmt.Sprintf(
			"https://iamcredentials.googleapis.com/v1/projects/-/serviceAccounts/%s:generateAccessToken",
			req.ServiceAccountEmail,
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

	return &GARAccessToken{Username: "oauth2accesstoken", Password: token.AccessToken}, nil
}
