package app

import "testing"

func TestAppBranchRunConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  AppBranchRunConfig
		wantErr bool
	}{
		{name: "push", config: AppBranchRunConfig{Mode: AppBranchRunModePush}},
		{name: "legacy all", config: AppBranchRunConfig{Mode: "all"}},
		{name: "manual only", config: AppBranchRunConfig{Mode: AppBranchRunModeManualOnly}},
		{name: "tag prefix", config: AppBranchRunConfig{Mode: AppBranchRunModeTagPrefix, TagPrefix: "foobar/"}},
		{name: "github label", config: AppBranchRunConfig{Mode: AppBranchRunModeGithubLabel, GithubLabel: "deploy-cadence-daily"}},
		{name: "tag prefix missing", config: AppBranchRunConfig{Mode: AppBranchRunModeTagPrefix}, wantErr: true},
		{name: "label missing", config: AppBranchRunConfig{Mode: AppBranchRunModeGithubLabel}, wantErr: true},
		{name: "field on push", config: AppBranchRunConfig{Mode: AppBranchRunModePush, TagPrefix: "foobar/"}, wantErr: true},
		{name: "unknown", config: AppBranchRunConfig{Mode: "sometimes"}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.Validate()
			if (err != nil) != test.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

func TestAppBranchRunConfigNormalize(t *testing.T) {
	for _, mode := range []AppBranchRunMode{"", "all"} {
		config := AppBranchRunConfig{Mode: mode}
		config.Normalize()
		if config.Mode != AppBranchRunModePush {
			t.Fatalf("Normalize() mode = %q, want %q", config.Mode, AppBranchRunModePush)
		}
	}
}
