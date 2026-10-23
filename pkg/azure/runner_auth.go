package azure

import "regexp"

// ManagementAudience is the Azure Resource Manager audience. A token for it is a
// live ARM credential for whatever identity minted it.
const ManagementAudience = "https://management.azure.com/"

// A Nuon app registration audience. Tokens for it cannot call ARM, and another
// user in the tenant cannot mint one until that tenant consents to the app.
var runnerAuthAudience = regexp.MustCompile(`^api://[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// ConfiguredRunnerAudience returns a Nuon app audience, or empty when the value
// is unset or not an api://<client-id> audience.
func ConfiguredRunnerAudience(configured string) string {
	if runnerAuthAudience.MatchString(configured) {
		return configured
	}
	return ""
}

// AuthAudiences is what ctl-api accepts. The ARM audience stays until every live
// runner requests the Nuon audience, because dropping it first logs them out.
func AuthAudiences(configured string) []string {
	audiences := []string{ManagementAudience}
	if extra := ConfiguredRunnerAudience(configured); extra != "" {
		audiences = append(audiences, extra)
	}
	return audiences
}

// RunnerAuthResource is the IMDS resource a runner should request. An empty
// value keeps the ARM audience. Anything else has to be that audience or a
// Nuon app audience.
func RunnerAuthResource(configured string) (string, bool) {
	if configured == "" || configured == ManagementAudience {
		return ManagementAudience, true
	}
	if extra := ConfiguredRunnerAudience(configured); extra != "" {
		return extra, true
	}
	return "", false
}
