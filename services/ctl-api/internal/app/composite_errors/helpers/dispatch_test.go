package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	composite_error "github.com/nuonco/nuon/pkg/composite_error"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestProducerFromComponentType(t *testing.T) {
	cases := []struct {
		in   app.ComponentType
		want string
	}{
		{app.ComponentTypeTerraformModule, "terraform"},
		{app.ComponentTypeHelmChart, "helm"},
		{app.ComponentTypeKubernetesManifest, "kubernetes"},
		{app.ComponentTypePulumi, "pulumi"},
		{app.ComponentTypeDockerBuild, ""},
		{app.ComponentTypeExternalImage, ""},
		{app.ComponentTypeJob, ""},
		{app.ComponentTypeUnknown, ""},
		{app.ComponentType("garbage"), ""},
	}
	for _, c := range cases {
		t.Run(string(c.in), func(t *testing.T) {
			assert.Equal(t, c.want, producerFromComponentType(c.in))
		})
	}
}

func TestPhaseFromTerraformStepName(t *testing.T) {
	cases := map[string]string{
		"terraform-apply":           "apply",
		"terraform-apply (plan)":    "plan",
		"terraform-apply (destroy)": "destroy",
		"TERRAFORM-APPLY (PLAN)":    "plan", // case-insensitive
		"":                          "apply",
		"some-step":                 "apply",
	}
	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, want, phaseFromTerraformStepName(name))
		})
	}
}

func TestPhaseFromStepName(t *testing.T) {
	cases := []struct {
		name     string
		producer string
		want     string
	}{
		{"helm-install", "helm", "install"},
		{"k8s-rollout", "kubernetes", "apply"},
		{"pulumi-up (plan)", "pulumi", "plan"},
		{"pulumi-up", "pulumi", "apply"},
		{"terraform-apply", "terraform", "apply"},
		{"terraform-apply (destroy)", "terraform", "destroy"},
		{"any", "", ""},        // unknown producer → empty
		{"any", "ansible", ""}, // unknown producer → empty
	}
	for _, c := range cases {
		t.Run(c.name+"|"+c.producer, func(t *testing.T) {
			assert.Equal(t, c.want, phaseFromStepName(c.name, c.producer))
		})
	}
}

func TestBuildParseContext(t *testing.T) {
	cases := []struct {
		producer, phase, cloud string
		want                   composite_error.ParseContext
	}{
		// happy path: full path
		{"terraform", "apply", "aws", "terraform/apply/aws"},
		{"helm", "install", "azure", "helm/install/azure"},

		// no cloud → drops cloud segment
		{"terraform", "plan", "", "terraform/plan"},

		// CloudPlatformUnknown is filtered out
		{"terraform", "apply", string(app.CloudPlatformUnknown), "terraform/apply"},

		// no phase → producer-only (cloud also dropped, since the chain
		// stops at the first empty segment)
		{"terraform", "", "aws", "terraform"},
		{"terraform", "", "", "terraform"},

		// no producer → empty (pipeline falls back to unknown_error)
		{"", "apply", "aws", ""},
		{"", "", "", ""},
	}
	for _, c := range cases {
		t.Run(string(c.want), func(t *testing.T) {
			assert.Equal(t, c.want, buildParseContext(c.producer, c.phase, c.cloud))
		})
	}
}
