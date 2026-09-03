package deployerrors

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func TestInstallStackOutdatedError(t *testing.T) {
	errorData, err := compositeerrors.New(&InstallStackOutdatedError{InstallID: "inl123"})
	require.NoError(t, err)
	require.Equal(t, InstallStackOutdatedErrorType, errorData.Type)
	require.Equal(t, compositeerrors.SeverityWarning, errorData.Severity)
	require.Equal(t, "Install stack is out of date", errorData.Message)
	require.Equal(t, "https://docs.nuon.co/guides/reprovision-installs", errorData.Hints.DocsURL())
	require.Len(t, errorData.Sections, 3)
	require.Contains(t, errorData.Sections[0].Body, "does not have an applied stack")
	require.Contains(t, errorData.Sections[1].Body, "**Manage**")
	require.Contains(t, errorData.Sections[1].Body, "**await install stack**")
	require.Equal(t, compositeerrors.SectionCode, errorData.Sections[2].Kind)
	require.Equal(t, "nuon installs stacks reprovision --install-id inl123", errorData.Sections[2].Body)
}
