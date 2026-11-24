package apps

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func TestAllowedHostsPolicyFromResponse(t *testing.T) {
	policy, err := allowedHostsPolicyFromResponse(&models.ServiceAppAllowedHosts{
		Mode: "warn",
		Categories: map[string]models.ServiceAppAllowedHostsCategory{
			"runner_api":      {Restricted: true, Entries: []string{"https://runner.example.com/"}},
			"stack_templates": {Restricted: false, Entries: []string{}},
			"not_known_yet":   {Restricted: true, Entries: []string{"https://example.com/"}},
		},
	})
	require.NoError(t, err)
	require.Equal(t, urlpolicy.ModeWarn, policy.Mode())
	require.True(t, policy.Restricted(urlpolicy.CategoryRunnerAPI))
	require.False(t, policy.Restricted(urlpolicy.CategoryStackTemplates))
	require.NotNil(t, policy.Check(urlpolicy.CategoryRunnerAPI, "runner.runner_api_url", "https://evil.test"))

	// A control plane with nothing restricted yields a policy that checks nothing.
	policy, err = allowedHostsPolicyFromResponse(&models.ServiceAppAllowedHosts{
		Mode: "enforce",
		Categories: map[string]models.ServiceAppAllowedHostsCategory{
			"runner_api": {Entries: []string{}},
		},
	})
	require.NoError(t, err)
	require.False(t, policy.AnyRestricted())

	policy, err = allowedHostsPolicyFromResponse(nil)
	require.NoError(t, err)
	require.False(t, policy.AnyRestricted())
}
