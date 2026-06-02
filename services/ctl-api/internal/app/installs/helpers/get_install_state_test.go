package helpers

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func cfgWithSubdomainInput() *app.AppConfig {
	return &app.AppConfig{
		InputConfig: app.AppInputConfig{
			AppInputs: []app.AppInput{
				{Name: "subdomain", Default: "whoami"},
				{Name: "name", Default: "whoami"},
			},
		},
	}
}

func strPtr(s string) *string { return &s }

// TestToInputState_AppliesDefaultsWhenNoValuesSet reproduces the reported bug:
// an install that has an inputs record but no values set produced a nil
// InputsState, so the app config's input defaults were never materialized and
// templates such as {{ .nuon.inputs.inputs.subdomain }} dereferenced a nil
// .nuon.inputs (panic: nil pointer evaluating interface {}.inputs).
func TestToInputState_AppliesDefaultsWhenNoValuesSet(t *testing.T) {
	cfg := cfgWithSubdomainInput()
	inputs := &app.InstallInputs{Values: pgtype.Hstore{}} // record exists, zero values set

	got := ToInputState(inputs, cfg, false)

	if got == nil {
		t.Fatal("ToInputState returned nil; expected a populated InputsState built from input defaults")
	}
	if got.Inputs["subdomain"] != "whoami" {
		t.Errorf("Inputs[\"subdomain\"] = %q; want %q", got.Inputs["subdomain"], "whoami")
	}
	if got.Inputs["name"] != "whoami" {
		t.Errorf("Inputs[\"name\"] = %q; want %q", got.Inputs["name"], "whoami")
	}
}

// TestToInputState_OverlaysSetValuesOnDefaults guards the existing behavior:
// values explicitly set on the install override the config defaults, while
// unset inputs still fall back to their defaults.
func TestToInputState_OverlaysSetValuesOnDefaults(t *testing.T) {
	cfg := cfgWithSubdomainInput()
	inputs := &app.InstallInputs{Values: pgtype.Hstore{"subdomain": strPtr("custom")}}

	got := ToInputState(inputs, cfg, false)

	if got == nil {
		t.Fatal("ToInputState returned nil; expected an InputsState")
	}
	if got.Inputs["subdomain"] != "custom" {
		t.Errorf("Inputs[\"subdomain\"] = %q; want %q", got.Inputs["subdomain"], "custom")
	}
	if got.Inputs["name"] != "whoami" {
		t.Errorf("Inputs[\"name\"] = %q; want %q (default)", got.Inputs["name"], "whoami")
	}
}
