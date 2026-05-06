package runnerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestApplyDigestPin(t *testing.T) {
	tests := []struct {
		name     string
		runner   *app.Runner
		settings app.RunnerGroupSettings
		wantTag  string
	}{
		{
			name:     "no digest leaves tag unchanged",
			runner:   &app.Runner{},
			settings: app.RunnerGroupSettings{ContainerImageTag: "prod"},
			wantTag:  "prod",
		},
		{
			name:     "digest is appended with @ separator",
			runner:   &app.Runner{ContainerImageDigest: "sha256:abc123"},
			settings: app.RunnerGroupSettings{ContainerImageTag: "prod"},
			wantTag:  "prod@sha256:abc123",
		},
		{
			name:     "tag already containing digest is left alone (idempotent)",
			runner:   &app.Runner{ContainerImageDigest: "sha256:abc123"},
			settings: app.RunnerGroupSettings{ContainerImageTag: "prod@sha256:def456"},
			wantTag:  "prod@sha256:def456",
		},
		{
			name:     "whitespace-only digest treated as unset",
			runner:   &app.Runner{ContainerImageDigest: "  "},
			settings: app.RunnerGroupSettings{ContainerImageTag: "prod"},
			wantTag:  "prod",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ApplyDigestPin(tc.runner, tc.settings)
			assert.Equal(t, tc.wantTag, got.ContainerImageTag)
		})
	}
}

func TestApplyDigestPinDoesNotMutateInput(t *testing.T) {
	runner := &app.Runner{ContainerImageDigest: "sha256:abc"}
	original := app.RunnerGroupSettings{ContainerImageTag: "prod"}
	_ = ApplyDigestPin(runner, original)
	assert.Equal(t, "prod", original.ContainerImageTag, "input settings must not be mutated")
}
