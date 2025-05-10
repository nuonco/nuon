package terraform

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube/secret"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/types/outputs"
)

func (p *handler) execSyncSecret(ctx context.Context, secr plantypes.KubernetesSecretSync) error {
	val, ts, _ := p.fetchSecret(ctx, secr)
	exists := val != ""

	if exists {
		if err := p.upsertSecret(ctx, secr, val); err != nil {
			return err
		}
	}

	p.state.outputs[secr.Name] = outputs.SecretSyncOutput{
		Name:                secr.Name,
		KubernetesNamespace: secr.Namespace,
		KubernetesName:      secr.Name,
		KubernetesKey:       secr.KeyName,
		Exists:              exists,

		Timestamp: ts,
		Length:    len(val),
	}

	return nil
}

func (p *handler) fetchSecret(ctx context.Context, secr plantypes.KubernetesSecretSync) (string, *time.Time, error) {
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
