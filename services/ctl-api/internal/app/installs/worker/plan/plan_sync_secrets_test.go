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

func ptr(v string) *string {
	return &v
}
