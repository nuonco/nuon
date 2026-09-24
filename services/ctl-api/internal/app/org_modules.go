package app

import "strings"

// ModuleFeaturePrefix is shared by every flag that hides a Lite dashboard
// module. The suffix is the module id the dashboard registry uses.
const ModuleFeaturePrefix = "disable-module-"

// ModuleFeatures lists the flags that hide Lite dashboard modules, in the
// order the dashboard's module registry declares them. Each one is off by
// default and admin managed: a vendor's bespoke control plane trims modules
// by pinning these through forced_enabled_features, and Nuon staff flip them
// per org from the Lite modules page.
func ModuleFeatures() []OrgFeature {
	return []OrgFeature{
		OrgFeatureDisableModuleApps,
		OrgFeatureDisableModuleInstalls,
		OrgFeatureDisableModuleTeam,
		OrgFeatureDisableModuleConnections,
		OrgFeatureDisableModuleWebhooks,
		OrgFeatureDisableModuleTriggers,
		OrgFeatureDisableModuleAPITokens,
		OrgFeatureDisableModuleServiceAccounts,
		OrgFeatureDisableModuleOIDCFederation,
	}
}

// ModuleID returns the dashboard module id a module flag hides, or "" when the
// flag is not a module flag.
func ModuleID(feature OrgFeature) string {
	name := string(feature)
	if !strings.HasPrefix(name, ModuleFeaturePrefix) {
		return ""
	}
	return strings.TrimPrefix(name, ModuleFeaturePrefix)
}
