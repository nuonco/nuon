package stack

import (
	"testing"

	"github.com/nuonco/nuon/sdks/stack/internal/core"
)

func TestApplyCloudOptionsGCP(t *testing.T) {
	i := &Installer{
		opts: Options{
			Cloud: core.CloudGCP,
			GCP:   GCPOptions{ProjectID: "my-proj", Region: "us-central1", RunnerMachineType: "e2-medium"},
		},
		// ctl-api-provided block carries the Nuon-generated fields.
		cfg: &Config{GCP: &core.GCPConfig{RunnerInitScriptURL: "https://example/init.sh"}},
	}

	i.applyCloudOptions()

	if i.cfg.Cloud != core.CloudGCP {
		t.Errorf("Cloud = %q, want gcp", i.cfg.Cloud)
	}
	if i.cfg.GCP.ProjectID != "my-proj" || i.cfg.GCP.Region != "us-central1" {
		t.Errorf("project/region not overlaid: %+v", i.cfg.GCP)
	}
	if i.cfg.GCP.RunnerMachineType != "e2-medium" {
		t.Errorf("machine type not overlaid: %q", i.cfg.GCP.RunnerMachineType)
	}
	// ctl-api-provided field must survive the overlay.
	if i.cfg.GCP.RunnerInitScriptURL != "https://example/init.sh" {
		t.Errorf("init script URL clobbered: %q", i.cfg.GCP.RunnerInitScriptURL)
	}
}

func TestApplyCloudOptionsAWSRegionFallback(t *testing.T) {
	i := &Installer{
		opts: Options{Cloud: core.CloudAWS, AWSRegion: "us-east-1"},
		cfg:  &Config{},
	}
	i.applyCloudOptions()
	if i.cfg.AWS == nil || i.cfg.AWS.Region != "us-east-1" {
		t.Errorf("aws region not applied: %+v", i.cfg.AWS)
	}
}

func TestValidateRequiredValues(t *testing.T) {
	i := &Installer{cfg: &Config{
		RequiredInputs: []string{"db_name"},
		InstallInputs:  map[string]string{"db_name": ""},
		Secrets: map[string]core.SecretInput{
			"api_key": {Required: true, Value: ""},
			"opt":     {Required: false, Value: ""},
		},
	}}
	if err := i.validateRequiredValues(); err == nil {
		t.Fatal("expected error when required input + secret are empty")
	}

	i.cfg.InstallInputs["db_name"] = "mydb"
	sec := i.cfg.Secrets["api_key"]
	sec.Value = "shh"
	i.cfg.Secrets["api_key"] = sec
	if err := i.validateRequiredValues(); err != nil {
		t.Errorf("unexpected error once required values are set: %v", err)
	}
}

func TestValidateLocation(t *testing.T) {
	cases := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{"aws ok", Options{Cloud: core.CloudAWS, AWSRegion: "us-east-1"}, false},
		{"aws missing region", Options{Cloud: core.CloudAWS}, true},
		// GCP location is collected after construction and validated at provision
		// time, so validateLocation does not require it here.
		{"gcp ok", Options{Cloud: core.CloudGCP, GCP: GCPOptions{ProjectID: "p", Region: "r"}}, false},
		{"gcp missing region ok at construction", Options{Cloud: core.CloudGCP, GCP: GCPOptions{ProjectID: "p"}}, false},
		{"default cloud aws", Options{AWSRegion: "us-east-1"}, false},
	}
	for _, c := range cases {
		_, _, err := validateLocation(c.opts)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v, wantErr = %v", c.name, err, c.wantErr)
		}
	}
}
