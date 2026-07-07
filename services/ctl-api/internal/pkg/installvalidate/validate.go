package installvalidate

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/appconfiggraph"
)

// OperationKind identifies the mutation being validated, which scopes how the
// rules render. A full sync evaluates the whole graph; an enable/disable scopes
// to the single component being toggled so an unrelated pre-existing problem
// elsewhere never blocks a targeted toggle.
type OperationKind int

const (
	OperationSync OperationKind = iota
	OperationEnable
	OperationDisable
)

// Operation describes the mutation being validated.
type Operation struct {
	Kind OperationKind
	// ComponentID is the component being toggled for OperationEnable/Disable.
	ComponentID string
}

// Context carries everything the rules need to evaluate an install's desired
// state: the dependency/enablement resolver, the component config snapshot, and
// the operation being validated.
type Context struct {
	Resolver *appconfiggraph.ComponentEnablementResolver
	CCCByID  map[string]*app.ComponentConfigConnection
	Op       Operation
}

// NewContext builds a validation context from the install's pinned app-config
// snapshot (component-ID keyed) and the install's input values, which carry the
// per-component enabled toggles.
func NewContext(cccByID map[string]*app.ComponentConfigConnection, enabledInputs map[string]*string, op Operation) *Context {
	return &Context{
		Resolver: appconfiggraph.NewComponentEnablementResolver(cccByID, enabledInputs),
		CCCByID:  cccByID,
		Op:       op,
	}
}

func (c *Context) name(id string) string {
	if ccc, ok := c.CCCByID[id]; ok && ccc != nil && ccc.Component.Name != "" {
		return ccc.Component.Name
	}
	return id
}

// Rule is a single validation check over the desired install state. New
// validation scenarios (beyond component toggles) are added by implementing
// Rule and registering it with a Validator.
type Rule interface {
	Name() string
	Check(*Context) Diagnostics
}

// Validator runs an ordered set of rules and aggregates their diagnostics.
type Validator struct {
	rules []Rule
}

func New(rules ...Rule) *Validator {
	return &Validator{rules: rules}
}

// DefaultValidator returns the validator with the standard install rules.
func DefaultValidator() *Validator {
	return New(
		EnableRequiresDependenciesEnabledRule{},
		DisableRequiresDependentsDisabledRule{},
	)
}

// Validate runs every rule and returns the deduplicated diagnostics.
func (v *Validator) Validate(c *Context) Diagnostics {
	var ds Diagnostics
	for _, r := range v.rules {
		ds = append(ds, r.Check(c)...)
	}
	return ds.dedup()
}
