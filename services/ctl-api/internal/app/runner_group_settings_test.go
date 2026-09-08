package app

import (
	"sync"
	"testing"

	"gorm.io/gorm/schema"
)

func TestRunnerGroupSettingsVendorTelemetryDefaultsToDisabled(t *testing.T) {
	model, err := schema.Parse(&RunnerGroupSettings{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	field := model.LookUpField("VendorTelemetryEnabled")
	if field == nil || !field.NotNull || !field.HasDefaultValue || field.DefaultValueInterface != false {
		t.Fatalf("vendor telemetry must have a non-null false database default: %+v", field)
	}
}
