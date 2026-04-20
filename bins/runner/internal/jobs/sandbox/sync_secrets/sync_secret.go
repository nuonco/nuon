package terraform

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/kube/secret"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/types/outputs"
)

func (p *handler) execSyncSecret(ctx context.Context, secr plantypes.KubernetesSecretSync) error {
	var val string
	var ts *time.Time

	switch {
	case secr.GCPSecretName != "":
		val, ts, _ = p.fetchGCPSecret(ctx, secr)
	case secr.AzureKeyVaultSecretID != "":
		val, ts, _ = p.fetchAzureSecret(ctx, secr)
	default:
		val, ts, _ = p.fetchAWSSecret(ctx, secr)
	}

	exists := val != ""

	if exists {
		if secr.Format == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(val)
			if err != nil {
				return errors.Wrap(err, "unable to decode base64 secret value")
			}
			val = strings.TrimSpace(string(decoded))
		}

		if err := p.upsertSecret(ctx, secr, val); err != nil {
			return err
		}
	}

	output := outputs.SecretSyncOutput{
		Name:                secr.Name,
		KubernetesNamespace: secr.Namespace,
		KubernetesName:      secr.Name,
		KubernetesKey:       secr.KeyName,
		Exists:              exists,
		Timestamp:           ts,
		Length:              len(val),
	}
	switch {
	case secr.GCPSecretName != "":
		output.GCPSecretName = secr.GCPSecretName
	case secr.AzureKeyVaultSecretID != "":
		output.AzureKeyVaultSecretID = secr.AzureKeyVaultSecretID
	default:
		output.ARN = secr.SecretARN
	}

	p.state.outputs[secr.Name] = output

	return nil
}

func (p *handler) fetchAWSSecret(ctx context.Context, secr plantypes.KubernetesSecretSync) (string, *time.Time, error) {
	cfg, err := awscredentials.Fetch(ctx, p.state.plan.AWSAuth)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get aws credentials")
	}

	svc := secretsmanager.NewFromConfig(cfg)

	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secr.SecretARN),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}

	result, err := svc.GetSecretValue(context.TODO(), input)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get latest value of secret")
	}

	return generics.FromPtrStr(result.SecretString), result.CreatedDate, nil
}

func (p *handler) fetchGCPSecret(ctx context.Context, secr plantypes.KubernetesSecretSync) (string, *time.Time, error) {
	gcpAuth := p.state.plan.GCPAuth

	var opts []option.ClientOption
	if gcpAuth != nil && gcpAuth.ImpersonateServiceAccount != "" {
		ts, err := impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
			TargetPrincipal: gcpAuth.ImpersonateServiceAccount,
			Scopes:          []string{"https://www.googleapis.com/auth/cloud-platform"},
		})
		if err != nil {
			return "", nil, errors.Wrap(err, "unable to create impersonated credentials")
		}
		opts = append(opts, option.WithTokenSource(ts))
	}

	client, err := secretmanager.NewClient(ctx, opts...)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to create secret manager client")
	}
	defer client.Close()

	result, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secr.GCPSecretName,
	})
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to access secret version")
	}

	return string(result.Payload.Data), nil, nil
}

func (p *handler) fetchAzureSecret(ctx context.Context, secr plantypes.KubernetesSecretSync) (string, *time.Time, error) {
	// Parse the secret URI to extract vault URL and secret name
	// Format: https://{vault-name}.vault.azure.net/secrets/{secret-name}
	parsed, err := url.Parse(secr.AzureKeyVaultSecretID)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to parse azure key vault secret URI")
	}

	vaultURL := parsed.Scheme + "://" + parsed.Host
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "secrets" {
		return "", nil, errors.Errorf("invalid key vault secret URI path: %s", parsed.Path)
	}
	secretName := parts[1]

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to create azure credentials")
	}

	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to create key vault client")
	}

	resp, err := client.GetSecret(ctx, secretName, "", nil)
	if err != nil {
		return "", nil, errors.Wrap(err, "unable to get secret from key vault")
	}

	val := ""
	if resp.Value != nil {
		val = *resp.Value
	}

	var ts *time.Time
	if resp.Attributes != nil && resp.Attributes.Updated != nil {
		ts = resp.Attributes.Updated
	}

	return val, ts, nil
}

func (p *handler) upsertSecret(ctx context.Context, secr plantypes.KubernetesSecretSync, val string) error {
	secrMgr, err := secret.New(p.v,
		secret.WithCluster(p.state.plan.ClusterInfo),
		secret.WithName(secr.Name),
		secret.WithNamespace(secr.Namespace),
		secret.WithKey(secr.KeyName),
	)
	if err != nil {
		return errors.Wrap(err, "unable to create secret manager")
	}

	if err := secrMgr.Upsert(ctx, []byte(val)); err != nil {
		return errors.Wrap(err, "unable to upsert secret")
	}

	return nil
}
