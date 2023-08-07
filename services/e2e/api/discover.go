package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/e2e/api/internal/introspection"
)

type discoverEndpoint struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

type discoverResponse struct {
	Description string             `json:"description"`
	Endpoints   []discoverEndpoint `json:"endpoints"`
}

func discoverHandler(ctx *gin.Context) {
	resp := &discoverResponse{
		Description: "This api exposes introspection details of a Nuon app running in a customer's cloud account.",
		Endpoints: []discoverEndpoint{
			{
				Description: introspection.KubeDescription,
				Path:        "/introspect/kube",
			},
			{
				Description: introspection.KubeNamespaceDescription,
				Path:        "/introspect/namespace/:namespace",
			},
			{
				Description: introspection.HelmDescription,
				Path:        "/introspect/helm",
			},
			{
				Description: introspection.HelmValuesDescription,
				Path:        "/introspect/helm-values/:namespace/:name",
			},
			{
				Description: introspection.HelmRenderedDescription,
				Path:        "/introspect/helm-rendered/:namespace/:name",
			},
			{
				Description: introspection.EnvDescription,
				Path:        "/introspect/env",
			},
			{
				Description: introspection.TerraformDescription,
				Path:        "/introspect/terraform",
			},
			{
				Description: introspection.SecretsDescription,
				Path:        "/introspect/secrets",
			},
			{
				Description: introspection.DefaultsDescription,
				Path:        "/introspect/defaults",
			},
			{
				Description: introspection.SandboxDescription,
				Path:        "/introspect/sandbox",
			},
			{
				Description: introspection.NuonDescription,
				Path:        "/introspect/nuon",
			},
			{
				Description: introspection.DockerBuildDescription,
				Path:        "/introspect/docker-build",
			},
			{
				Description: introspection.ExternalImageDescription,
				Path:        "/introspect/external-image",
			},
			{
				Description: "/livez check",
				Path:        "/livez",
			},
			{
				Description: "/readyz check",
				Path:        "/readyz",
			},
		},
	}

	ctx.JSON(http.StatusOK, resp)
}
