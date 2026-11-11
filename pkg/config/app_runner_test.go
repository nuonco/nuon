package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppRunnerConfig_Parse_ImageID(t *testing.T) {
	tests := map[string]struct {
		runnerType string
		imageID    string
		wantErr    bool
	}{
		"unset on aws":            {runnerType: "aws"},
		"valid ami on aws":        {runnerType: "aws", imageID: "ami-0abcdef1234567890"},
		"valid ami on aws-eks":    {runnerType: "aws-eks", imageID: "ami-0abc123"},
		"missing prefix on aws":   {runnerType: "aws", imageID: "0abcdef1234567890", wantErr: true},
		"uppercase hex on aws":    {runnerType: "aws", imageID: "ami-0ABCDEF", wantErr: true},
		"trailing junk on aws":    {runnerType: "aws", imageID: "ami-0abc123!", wantErr: true},
		"ssm alias on aws":        {runnerType: "aws", imageID: "resolve:ssm:/aws/service/ami", wantErr: true},
		"non-ami format on gcp":   {runnerType: "gcp", imageID: "projects/acme/global/images/runner"},
		"non-ami format on azure": {runnerType: "azure", imageID: "Canonical:ubuntu:22_04-lts-gen2:latest"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			cfg := &AppRunnerConfig{RunnerType: tc.runnerType, ImageID: tc.imageID}

			err := cfg.parse(ParseOptions{})

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
