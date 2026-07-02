// Package errparse turns the raw error output of a failed runner job into a
// typed compositeerrors.CompositeError.
//
// Parsing is hierarchical: a single failure blob is a stack of nested errors
// (cloud-provider inside tool inside orchestration), and we want to surface the
// most specific layer — "Missing IAM permission s3:CreateBucket", not
// "terraform apply failed". Parsers therefore declare a Layer, and the registry
// tries the deepest (provider) layer first, then tool, then a generic
// emit-the-body fallback.
//
// The deepest layer holds the most parsers and grows without bound, so gating
// must be cheap: a parser advertises the Tools it applies to (a free facet
// index) and the Signals that must be physically present in the text (a single
// compiled multi-pattern scan). A parser's expensive Parse only runs when its
// facet gate passes and one of its signals is present, so per-error cost is a
// function of the text length and the handful of matched candidates — not the
// total number of registered parsers.
package errparse

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// Layer orders parsers from most-specific to generic. Lower layers are tried
// first, so a provider-level cause wins over the tool-level fallback.
type Layer int

const (
	// LayerProvider is the cloud-provider layer (AWS/Azure/GCP error codes).
	LayerProvider Layer = 0
	// LayerTool is the execution-tool layer (terraform/helm/pulumi/docker).
	LayerTool Layer = 10
	// LayerGeneric is the always-matches fallback that emits the cleaned error
	// body when no specific parser recognises the failure.
	LayerGeneric Layer = 100
)

// Tool is the underlying execution tool a runner job used. It is derived from
// the runner job type and used as a free facet to bucket parsers, so terraform
// parsers are never even considered for a helm job.
type Tool string

const (
	ToolUnknown    Tool = ""
	ToolTerraform  Tool = "terraform"
	ToolHelm       Tool = "helm"
	ToolPulumi     Tool = "pulumi"
	ToolDocker     Tool = "docker"
	ToolKubernetes Tool = "kubernetes"
	ToolOCI        Tool = "oci"
)

// Provider is the cloud provider a job targeted. It is resolved lazily because
// it costs a DB lookup, and only matters once a provider-specific signal has
// already been found in the text.
type Provider string

const (
	ProviderUnknown Provider = ""
	ProviderAWS     Provider = "aws"
	ProviderAzure   Provider = "azure"
	ProviderGCP     Provider = "gcp"
)

// Owner identifies the row a runner job belongs to (polymorphic: table name +
// id), e.g. ("install_deploys", "<id>").
type Owner struct {
	Type string
	ID   string
}

// ParseContext carries the facets a parser gates and parses on. The free facets
// (Raw, Tool, Operation, Group, Owner, Meta) are populated eagerly by the
// caller from the already-loaded runner job; Provider is resolved lazily and
// memoized so an unmatched job never pays for the DB lookup.
type ParseContext struct {
	// Raw is the untruncated error output captured from the job.
	Raw string

	// Tool is the execution tool, or ToolUnknown when it can't be determined.
	Tool Tool

	// Operation is the job operation (e.g. "apply-plan"), free-form.
	Operation string

	// Group is the runner job group (e.g. "deploy", "build", "sandbox").
	Group string

	// Owner is the row the job belongs to.
	Owner Owner

	// Meta carries the runner-sent error metadata (step, handler, job_type, ...).
	Meta map[string]string

	// ResolveProvider lazily resolves the cloud provider. It is optional: when
	// nil, Provider() reports ProviderUnknown.
	ResolveProvider func() Provider

	provider     Provider
	providerDone bool
}

// Provider returns the cloud provider, resolving it at most once. It fails open
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

// Parser recognises one class of failure. Implementations must be pure and
// local: no DB or network access (the one lazy DB facet, Provider, is resolved
// by the context, not the parser).
type Parser interface {
	// Layer places the parser in the specificity hierarchy.
	Layer() Layer

	// Tools are the execution tools this parser applies to. An empty slice
	// means tool-agnostic (considered for every job). Used to bucket parsers
	// for cheap facet-based narrowing.
	Tools() []Tool

	// Signals are substrings that must be present in the raw text for this
	// parser to be a candidate. They are compiled into a single multi-pattern
	// matcher per bucket. An empty slice means "always a candidate" (used by
	// the generic fallback), which never runs the matcher.
	Signals() []string

	// Applicable confirms the parser should run after a signal match, using the
	// finer facets (notably Provider). It must fail open on unknown facets.
	Applicable(ctx *ParseContext) bool

	// Parse attempts to produce a typed error, or returns nil to defer to the
	// next candidate. It must be selective: a miss returns nil rather than a
	// low-confidence match.
	Parse(ctx *ParseContext) compositeerrors.CompositeError
}
