package installvalidate

import (
	"strconv"
	"testing"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/refs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func toggleable(id, name string, defaultEnabled bool, depIDs ...string) *app.ComponentConfigConnection {
	tr := true
	de := defaultEnabled
	return &app.ComponentConfigConnection{
		ComponentID:            id,
		Component:              app.Component{ID: id, Name: name},
		Toggleable:             &tr,
		DefaultEnabled:         &de,
		ComponentDependencyIDs: pq.StringArray(depIDs),
	}
}

func plain(id, name string, depIDs ...string) *app.ComponentConfigConnection {
	return &app.ComponentConfigConnection{
		ComponentID:            id,
		Component:              app.Component{ID: id, Name: name},
		ComponentDependencyIDs: pq.StringArray(depIDs),
	}
}

// ctx builds a validation context, materializing per-component-name toggles as
// the reserved synthetic enabled inputs. In these tests component name == id.
func ctx(toggles map[string]bool, op Operation, cccs ...*app.ComponentConfigConnection) *Context {
	byID := make(map[string]*app.ComponentConfigConnection, len(cccs))
	for _, c := range cccs {
		byID[c.ComponentID] = c
	}
	inputs := make(map[string]*string, len(toggles))
	for name, enabled := range toggles {
		v := strconv.FormatBool(enabled)
		inputs[config.EnabledOverrideInputName(name)] = &v
	}
	return NewContext(byID, inputs, op)
}

func TestSync_ConsistentState_NoDiagnostics(t *testing.T) {
	c := ctx(map[string]bool{"a": true},
		Operation{Kind: OperationSync},
		toggleable("a", "a", true),
		plain("b", "b", "a"),
	)
	ds := DefaultValidator().Validate(c)
	assert.Empty(t, ds)
}

func TestSync_EnabledDependentOnDisabledDep_OneErrorDeduped(t *testing.T) {
	c := ctx(map[string]bool{"a": false},
		Operation{Kind: OperationSync},
		toggleable("a", "a", true),
		plain("b", "b", "a"),
	)
	ds := DefaultValidator().Validate(c)

	assert.True(t, ds.HasErrors())
	assert.Len(t, ds, 1, "the single invalid edge must be reported once, not once per rule")
	assert.Equal(t, CodeDisabledDependency, ds[0].Code)
	assert.ElementsMatch(t, []string{"a", "b"}, ds[0].Components)
}

func TestSync_CascadeViaOutputRef(t *testing.T) {
	b := plain("b", "b")
	b.Refs = []refs.Ref{{Type: refs.RefTypeComponents, Name: "a", Value: "url"}}
	c := ctx(map[string]bool{"a": false},
		Operation{Kind: OperationSync},
		toggleable("a", "a", true),
		b,
	)
	ds := DefaultValidator().Validate(c)
	assert.Len(t, ds, 1)
}

func TestEnable_TargetWithDisabledDep_Errors(t *testing.T) {
	c := ctx(map[string]bool{"a": false, "b": true},
		Operation{Kind: OperationEnable, ComponentID: "b"},
		toggleable("a", "a", true),
		toggleable("b", "b", true, "a"),
	)
	ds := DefaultValidator().Validate(c)
	assert.True(t, ds.HasErrors())
	assert.Len(t, ds, 1)
}

func TestEnable_TargetWithEnabledDep_OK(t *testing.T) {
	c := ctx(map[string]bool{"a": true, "b": true},
		Operation{Kind: OperationEnable, ComponentID: "b"},
		toggleable("a", "a", true),
		toggleable("b", "b", true, "a"),
	)
	ds := DefaultValidator().Validate(c)
	assert.Empty(t, ds)
}

func TestEnable_IgnoresUnrelatedInvalidEdge(t *testing.T) {
	// c<-d is an unrelated pre-existing invalid edge; enabling b (whose dep a is
	// enabled) must not be blocked by it.
	c := ctx(map[string]bool{"a": true, "b": true, "cc": false, "d": true},
		Operation{Kind: OperationEnable, ComponentID: "b"},
		toggleable("a", "a", true),
		toggleable("b", "b", true, "a"),
		toggleable("cc", "cc", true),
		toggleable("d", "d", true, "cc"),
	)
	ds := DefaultValidator().Validate(c)
	assert.Empty(t, ds, "enable b must not surface the unrelated cc/d problem")
}

func TestDisable_TargetWithEnabledDependent_Errors(t *testing.T) {
	c := ctx(map[string]bool{"a": false, "b": true},
		Operation{Kind: OperationDisable, ComponentID: "a"},
		toggleable("a", "a", true),
		toggleable("b", "b", true, "a"),
	)
	ds := DefaultValidator().Validate(c)
	assert.True(t, ds.HasErrors())
	assert.Len(t, ds, 1)
	assert.ElementsMatch(t, []string{"a", "b"}, ds[0].Components)
}

func TestDisable_TargetWithDisabledDependent_OK(t *testing.T) {
	c := ctx(map[string]bool{"a": false, "b": false},
		Operation{Kind: OperationDisable, ComponentID: "a"},
		toggleable("a", "a", true),
		toggleable("b", "b", true, "a"),
	)
	ds := DefaultValidator().Validate(c)
	assert.Empty(t, ds)
}

func TestErrorAggregation(t *testing.T) {
	ds := Diagnostics{
		{Severity: SeverityError, Summary: "first"},
		{Severity: SeverityError, Summary: "second"},
	}
	assert.Equal(t, "first; second", ds.Error())
}
