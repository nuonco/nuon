package config

import (
	"encoding/hex"
	"strings"
)

// ComponentOverrideInputPrefix is the reserved prefix for auto-generated
// per-component install-level override inputs (Helm values / Terraform vars).
//
// Synthetic input names have the shape:
//
//	nuon_component_override_v1_<kind>_<hex(componentName)>
//
// The hex encoding of the component name keeps the key TOML-safe ([a-z0-9_]),
// reversible, and collision-proof (e.g. "foo-bar" and "foo_bar" cannot collide).
// It is also a single segment with no dots/whitespace, so it matches the
// nuon.install.inputs.<name> reference regex used for redeploy-on-edit.
//
// Vendors are not permitted to declare inputs with this prefix.
const ComponentOverrideInputPrefix = "nuon_component_override_v1_"

// ComponentOverrideInputGroup is the reserved input group that all synthetic
// component-override inputs belong to. It is created automatically during app
// sync and is not meant to be referenced directly by vendors.
const ComponentOverrideInputGroup = "nuon_component_overrides"

// componentOverrideInputIndexBase offsets synthetic input indices well past any
// realistic vendor-declared input count, so they sort last and never collide
// with (or take the reserved zero value of) real input indices.
const componentOverrideInputIndexBase = 1_000_000

// ComponentOverrideInputGroupIndex orders the reserved override group after any
// realistic vendor-declared group.
const ComponentOverrideInputGroupIndex = 1_000_000

// ComponentOverrideKind identifies which override axis a synthetic input drives.
type ComponentOverrideKind string

const (
	ComponentOverrideKindHelmValues ComponentOverrideKind = "helm_values"
	ComponentOverrideKindTFVars     ComponentOverrideKind = "tf_vars"
)

// InputType returns the dedicated AppInput type used for this override kind's
// synthetic input, so its value is syntax-validated as structured data rather
// than treated as an opaque string.
func (k ComponentOverrideKind) InputType() string {
	switch k {
	case ComponentOverrideKindHelmValues:
		return InputTypeYAML
	case ComponentOverrideKindTFVars:
		return InputTypeHCL
	}
	return "string"
}

// componentOverrideInputName builds the reserved synthetic input name for a
// component override of the given kind.
func componentOverrideInputName(kind ComponentOverrideKind, componentName string) string {
	return ComponentOverrideInputPrefix + string(kind) + "_" + hex.EncodeToString([]byte(componentName))
}

// HelmValuesOverrideInputName returns the reserved synthetic input name that
// carries the install-level Helm values override for the named component.
func HelmValuesOverrideInputName(componentName string) string {
	return componentOverrideInputName(ComponentOverrideKindHelmValues, componentName)
}

// TFVarsOverrideInputName returns the reserved synthetic input name that carries
// the install-level Terraform vars override for the named component.
func TFVarsOverrideInputName(componentName string) string {
	return componentOverrideInputName(ComponentOverrideKindTFVars, componentName)
}

// IsComponentOverrideInputName reports whether an input name is a reserved
// component-override synthetic input.
func IsComponentOverrideInputName(name string) bool {
	return strings.HasPrefix(name, ComponentOverrideInputPrefix)
}

// SyntheticOverrideInput describes a single synthetic vendor input that must be
// declared on the app so a per-component install-level override can flow through
// the regular install-input machinery (validation, storage, state, redeploy).
type SyntheticOverrideInput struct {
	// Name is the reserved synthetic input name (see ComponentOverrideInputPrefix).
	Name string
	// Kind is the override axis this input drives (helm_values / tf_vars).
	Kind ComponentOverrideKind
	// Component is the component name this override targets.
	Component string
	// Index is the suggested ordering index, offset well past real inputs.
	Index int
}

// SyntheticComponentOverrideInputs enumerates the synthetic override inputs that
// must be declared for the given components: one helm_values input per Helm
// component and one tf_vars input per Terraform module component. The result is
// deterministic (ordered by component appearance, then kind) so repeated syncs
// produce identical app input configs.
func SyntheticComponentOverrideInputs(components ComponentList) []SyntheticOverrideInput {
	out := make([]SyntheticOverrideInput, 0, len(components))
	idx := componentOverrideInputIndexBase
	for _, comp := range components {
		if comp == nil {
			continue
		}
		switch comp.Type {
		case HelmChartComponentType:
			out = append(out, SyntheticOverrideInput{
				Name:      HelmValuesOverrideInputName(comp.Name),
				Kind:      ComponentOverrideKindHelmValues,
				Component: comp.Name,
				Index:     idx,
			})
			idx++
		case TerraformModuleComponentType:
			out = append(out, SyntheticOverrideInput{
				Name:      TFVarsOverrideInputName(comp.Name),
				Kind:      ComponentOverrideKindTFVars,
				Component: comp.Name,
				Index:     idx,
			})
			idx++
		}
	}
	return out
}

// ParseComponentOverrideInputName decodes a reserved synthetic input name back
// into its override kind and component name. ok is false when name is not a
// component-override input or is malformed.
func ParseComponentOverrideInputName(name string) (kind ComponentOverrideKind, componentName string, ok bool) {
	if !IsComponentOverrideInputName(name) {
		return "", "", false
	}

	rest := strings.TrimPrefix(name, ComponentOverrideInputPrefix)

	// rest is "<kind>_<hex>"; kind values themselves contain underscores, so
	// match against the known kinds rather than splitting on "_".
	for _, k := range []ComponentOverrideKind{ComponentOverrideKindHelmValues, ComponentOverrideKindTFVars} {
		prefix := string(k) + "_"
		if !strings.HasPrefix(rest, prefix) {
			continue
		}
		encoded := strings.TrimPrefix(rest, prefix)
		decoded, err := hex.DecodeString(encoded)
		if err != nil {
			return "", "", false
		}
		return k, string(decoded), true
	}

	return "", "", false
}
