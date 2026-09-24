package app

import (
	"slices"
	"testing"
)

func TestModuleFeaturesShipEveryModuleUnlessHidden(t *testing.T) {
	registered := GetFeatures()
	defaults := DefaultFeatures()
	descriptions := GetFeatureDescriptions()
	manageable := GetUserManageableFeatures()

	if len(ModuleFeatures()) == 0 {
		t.Fatal("at least one module flag must be declared")
	}

	for _, feature := range ModuleFeatures() {
		if !slices.Contains(registered, feature) {
			t.Errorf("%s must be a registered org feature", feature)
		}
		if defaults[feature] {
			t.Errorf("%s must default to off so the module ships unless hidden", feature)
		}
		if descriptions[feature] == "" {
			t.Errorf("%s must carry a description for the admin dashboard", feature)
		}
		if slices.Contains(manageable, feature) {
			t.Errorf("%s must be admin managed, not user manageable", feature)
		}
		if ModuleID(feature) == "" {
			t.Errorf("%s must carry the %q prefix", feature, ModuleFeaturePrefix)
		}
	}
}

func TestModuleIDIgnoresOtherFlags(t *testing.T) {
	if got := ModuleID(OrgFeatureDisableAppSync); got != "" {
		t.Fatalf("expected no module id for %s, got %q", OrgFeatureDisableAppSync, got)
	}
	if got := ModuleID(OrgFeatureDisableModuleInstalls); got != "installs" {
		t.Fatalf("expected installs, got %q", got)
	}
}
