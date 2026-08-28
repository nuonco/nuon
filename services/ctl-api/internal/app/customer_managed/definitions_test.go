package customermanaged

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestCanonicalRunbookDefinitionIgnoresPersistenceIdentity(t *testing.T) {
	first := app.RunbookConfig{ID: "runbook-config-a", AppConfigID: "app-config-a", RunbookID: "runbook-a", Readme: "deploy the app"}
	second := app.RunbookConfig{ID: "runbook-config-b", AppConfigID: "app-config-b", RunbookID: "runbook-b", Readme: "deploy the app"}

	require.Equal(t, CanonicalRunbookDefinition(first, nil), CanonicalRunbookDefinition(second, nil))
}

func TestCanonicalRunbookDefinitionIncludesSemanticChanges(t *testing.T) {
	first := app.RunbookConfig{Readme: "deploy the app"}
	second := app.RunbookConfig{Readme: "deploy the updated app"}

	require.NotEqual(t, ObjectDigest(CanonicalRunbookDefinition(first, nil)), ObjectDigest(CanonicalRunbookDefinition(second, nil)))
}
