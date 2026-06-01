package plan

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestGetKubernetesSecret_RequiredSecretMissingARNReturnsError(t *testing.T) {
	p := &Planner{}

	_, ok, err := p.getKubernetesSecret(
		app.InstallStackOutputs{Data: pgtype.Hstore{}},
		app.AppSecretConfig{Name: "db", Required: true},
	)

	require.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "db_arn")
}

func TestGetKubernetesSecret_OptionalSecretMissingARNIsSkipped(t *testing.T) {
	p := &Planner{}

	_, ok, err := p.getKubernetesSecret(
		app.InstallStackOutputs{Data: pgtype.Hstore{}},
		app.AppSecretConfig{Name: "db", Required: false},
	)

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestGetKubernetesSecret_OptionalSecretEmptyARNIsSkipped(t *testing.T) {
	p := &Planner{}

	_, ok, err := p.getKubernetesSecret(
		app.InstallStackOutputs{
			Data: pgtype.Hstore{
				"db_arn": ptr("")},
		},
		app.AppSecretConfig{Name: "db", Required: false},
	)

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestGetKubernetesSecret_ReturnsSecretWhenARNPresent(t *testing.T) {
	p := &Planner{}

	secret, ok, err := p.getKubernetesSecret(
		app.InstallStackOutputs{
			Data: pgtype.Hstore{
				"db_arn": ptr("arn:aws:secretsmanager:us-west-2:123:secret:db"),
			},
		},
		app.AppSecretConfig{
			Name:                      "db",
			KubernetesSecretName:      "app-db",
			KubernetesSecretKey:       "password",
			KubernetesSecretNamespace: "default",
		},
	)

	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, "arn:aws:secretsmanager:us-west-2:123:secret:db", secret.SecretARN)
	assert.Equal(t, "db", secret.SecretName)
	assert.Equal(t, "app-db", secret.Name)
	assert.Equal(t, "password", secret.KeyName)
	assert.Equal(t, "default", secret.Namespace)
}

func TestGetKubernetesSecret_EmitsV2Targets(t *testing.T) {
	p := &Planner{}

	secret, ok, err := p.getKubernetesSecret(
		app.InstallStackOutputs{
			Data: pgtype.Hstore{
				"db_arn": ptr("arn:aws:secretsmanager:us-west-2:123:secret:db"),
			},
		},
		app.AppSecretConfig{
			Name: "db",
			KubernetesSyncTargets: []app.AppSecretKubernetesSyncTarget{
				{Namespaces: []string{"cloudprem", "datadog"}, Name: "datadog", Key: "api-key"},
				{Namespaces: []string{"datadog"}, Name: "datadog", Key: "app-key"},
			},
		},
	)

	require.NoError(t, err)
	assert.True(t, ok)
	require.Len(t, secret.Targets, 2)
	assert.Equal(t, []string{"cloudprem", "datadog"}, secret.Targets[0].Namespaces)
	assert.Equal(t, "datadog", secret.Targets[0].Name)
	assert.Equal(t, "api-key", secret.Targets[0].Key)
	assert.Equal(t, "app-key", secret.Targets[1].Key)
}

func TestGetKubernetesSecret_NoTargetsLeavesV2Nil(t *testing.T) {
	p := &Planner{}

	secret, ok, err := p.getKubernetesSecret(
		app.InstallStackOutputs{
			Data: pgtype.Hstore{
				"db_arn": ptr("arn:aws:secretsmanager:us-west-2:123:secret:db"),
			},
		},
		app.AppSecretConfig{
			Name:                      "db",
			KubernetesSecretName:      "app-db",
			KubernetesSecretKey:       "password",
			KubernetesSecretNamespace: "default",
		},
	)

	require.NoError(t, err)
	assert.True(t, ok)
	assert.Nil(t, secret.Targets)
}

func ptr(v string) *string {
	return &v
}
