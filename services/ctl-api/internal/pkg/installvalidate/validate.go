package installvalidate

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/appconfiggraph"
)

type OperationKind int

const (
	OperationSync OperationKind = iota
	OperationEnable
	OperationDisable
)

type Operation struct {
	Kind        OperationKind
	ComponentID string
}

type Context struct {
	Resolver *appconfiggraph.ComponentEnablementResolver
	CCCByID  map[string]*app.ComponentConfigConnection
	Op       Operation
}

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

type Rule interface {
	Name() string
	Check(*Context) Diagnostics
}

type Validator struct {
	rules []Rule
}

func New(rules ...Rule) *Validator {
	return &Validator{rules: rules}
}

func DefaultValidator() *Validator {
	return New(
		EnableRequiresDependenciesEnabledRule{},
		DisableRequiresDependentsDisabledRule{},
	)
}

func (v *Validator) Validate(c *Context) Diagnostics {
	var ds Diagnostics
	for _, r := range v.rules {
		ds = append(ds, r.Check(c)...)
	}
	return ds.dedup()
}
