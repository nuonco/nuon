package terraform

import (
	"context"
	"encoding/base64"
	"fmt"
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

// execSyncSecret resolves the shared source value once, then dispatches to the v1 (single destination) or v2
// (multiple targets, each fanning out across namespaces) sync path.
func (p *handler) execSyncSecret(ctx context.Context, secr plantypes.KubernetesSecretSync) error {
	val, ts, exists, err := p.fetchSecretValue(ctx, secr)
	if err != nil {
		return err
	}

	if len(secr.Targets) > 0 {
		return p.execSyncSecretV2(ctx, secr, val, ts, exists)
	}

	return p.execSyncSecretV1(ctx, secr, val, ts, exists)
}

// fetchSecretValue resolves the secret value from the cloud provider source shared across all destinations, applying
// base64 decoding when configured. exists is false when the source resolves to an empty value (an optional secret
// that was never populated), in which case the runner records the output but skips the Kubernetes upsert.
func (p *handler) fetchSecretValue(ctx context.Context, secr plantypes.KubernetesSecretSync) (string, *time.Time, bool, error) {
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
	if exists && secr.Format == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return "", nil, false, errors.Wrap(err, "unable to decode base64 secret value")
		}
		val = strings.TrimSpace(string(decoded))
	}

	return val, ts, exists, nil
}

// execSyncSecretV1 is the legacy single-destination path: it upserts one Kubernetes secret using the v1 namespace /
// name / key fields and records a single output.
func (p *handler) execSyncSecretV1(ctx context.Context, secr plantypes.KubernetesSecretSync, val string, ts *time.Time, exists bool) error {
	if exists {
		if err := p.upsertSecret(ctx, secr.Namespace, secr.Name, secr.KeyName, val); err != nil {
			return err
		}
	}

	p.recordOutput(secr, secr.Namespace, secr.Name, secr.KeyName, val, ts, exists)

	return nil
}

// execSyncSecretV2 fans the shared source value out across every target × namespace, upserting one Kubernetes secret
// per destination and recording one output per destination.
func (p *handler) execSyncSecretV2(ctx context.Context, secr plantypes.KubernetesSecretSync, val string, ts *time.Time, exists bool) error {
	for _, target := range secr.Targets {
		for _, namespace := range target.Namespaces {
			if exists {
				if err := p.upsertSecret(ctx, namespace, target.Name, target.Key, val); err != nil {
					return err
				}
			}

			p.recordOutput(secr, namespace, target.Name, target.Key, val, ts, exists)
		}
	}

	return nil
}

// recordOutput writes a per-destination output keyed uniquely by source secret name and Kubernetes destination, so v2
// destinations that share a name (across namespaces, or same namespace/name with different keys) don't overwrite each
// other in the output map.
func (p *handler) recordOutput(secr plantypes.KubernetesSecretSync, namespace, name, key, val string, ts *time.Time, exists bool) {
	output := outputs.SecretSyncOutput{
		Name:                secr.SecretName,
		KubernetesNamespace: namespace,
		KubernetesName:      name,
		KubernetesKey:       key,
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

	p.state.outputs[fmt.Sprintf("%s/%s/%s/%s", secr.SecretName, namespace, name, key)] = output
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

func (p *handler) upsertSecret(ctx context.Context, namespace, name, key, val string) error {
	secrMgr, err := secret.New(p.v,
		secret.WithCluster(p.state.plan.ClusterInfo),
		secret.WithName(name),
		secret.WithNamespace(namespace),
		secret.WithKey(key),
	)
	if err != nil {
		return errors.Wrap(err, "unable to create secret manager")
	}

	if err := secrMgr.Upsert(ctx, []byte(val)); err != nil {
		return errors.Wrap(err, "unable to upsert secret")
	}

	return nil
}
