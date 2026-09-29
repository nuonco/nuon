package errparse

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

type Layer int

const (
	LayerProvider     Layer = 0
	LayerToolSpecific Layer = 9
	LayerTool         Layer = 10
	LayerGeneric      Layer = 100
)

type Tool string

const (
	ToolUnknown    Tool = ""
	ToolAction     Tool = "action"
	ToolTerraform  Tool = "terraform"
	ToolHelm       Tool = "helm"
	ToolPulumi     Tool = "pulumi"
	ToolDocker     Tool = "docker"
	ToolKubernetes Tool = "kubernetes"
	ToolOCI        Tool = "oci"
)

type Provider string

const (
	ProviderUnknown Provider = ""
	ProviderAWS     Provider = "aws"
	ProviderAzure   Provider = "azure"
	ProviderGCP     Provider = "gcp"
)

type Owner struct {
	Type string
	ID   string
}

type ParseContext struct {
	Raw string

	Tool Tool

	Operation string

	Group string

	Owner Owner

	Meta map[string]string

	ResolveProvider func() Provider

	provider     Provider
	providerDone bool
}

// why: Provider returns the cloud provider, resolving it at most once. It fails open
// to ProviderUnknown so an unresolvable owner never suppresses provider parsing
// (provider gating is a false-positive guard, never the sole gate).
func (c *ParseContext) Provider() Provider {
	if c.providerDone {
		return c.provider
	}
	c.providerDone = true
	if c.ResolveProvider != nil {
		c.provider = c.ResolveProvider()
	}
	return c.provider
}

type Parser interface {
	Layer() Layer

	Tools() []Tool

	Signals() []string

	Applicable(ctx *ParseContext) bool

	Parse(ctx *ParseContext) compositeerrors.CompositeError
}
