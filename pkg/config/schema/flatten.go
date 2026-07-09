package schema

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/invopop/jsonschema"

	"github.com/nuonco/nuon/pkg/config"
)

// componentNestedConfigKeys are the Component properties that hold nested
// per-type config blocks. Per-type component files are decoded flattened (see
// config.DecodeComponent): the typed config's fields live at the top level of
// the file, so the nested keys — and Component's oneOf over them — do not
// apply to a single-type file schema.
var componentNestedConfigKeys = map[string]struct{}{
	"helm_chart":          {},
	"terraform_module":    {},
	"docker_build":        {},
	"job":                 {},
	"external_image":      {},
	"kubernetes_manifest": {},
	"pulumi":              {},
}

// flattenedComponentSchema builds the schema for a single-type component file
// by merging Component's common fields and the typed config's fields into one
// object schema. The previous allOf composition was unsatisfiable under strict
// JSON Schema semantics: each branch set additionalProperties: false, and
// additionalProperties cannot see properties declared in sibling branches, so
// every branch rejected the other branch's keys.
func flattenedComponentSchema(typedConfig any) (*jsonschema.Schema, error) {
	r, err := reflector()
	if err != nil {
		return nil, err
	}

	compRoot := r.Reflect(config.Component{})
	typedRoot := r.Reflect(typedConfig)

	comp, err := rootDefinition(compRoot)
	if err != nil {
		return nil, err
	}
	typed, err := rootDefinition(typedRoot)
	if err != nil {
		return nil, err
	}

	if err := checkUnhandledConstraints(comp, "config.Component"); err != nil {
		return nil, err
	}
	if err := checkUnhandledConstraints(typed, fmt.Sprintf("%T", typedConfig)); err != nil {
		return nil, err
	}

	merged := &jsonschema.Schema{
		Version:              compRoot.Version,
		Type:                 "object",
		Title:                typed.Title,
		Description:          typed.Description,
		Properties:           jsonschema.NewProperties(),
		AdditionalProperties: jsonschema.FalseSchema,
		Definitions:          jsonschema.Definitions{},
	}
	if merged.Title == "" {
		merged.Title = comp.Title
	}
	if merged.Description == "" {
		merged.Description = comp.Description
	}

	for pair := comp.Properties.Oldest(); pair != nil; pair = pair.Next() {
		if _, skip := componentNestedConfigKeys[pair.Key]; skip {
			continue
		}
		merged.Properties.Set(pair.Key, pair.Value)
	}
	for pair := typed.Properties.Oldest(); pair != nil; pair = pair.Next() {
		if _, exists := merged.Properties.Get(pair.Key); exists {
			return nil, fmt.Errorf("property %q is declared by both config.Component and %T", pair.Key, typedConfig)
		}
		merged.Properties.Set(pair.Key, pair.Value)
	}

	merged.Required = mergeRequired(comp.Required, typed.Required)
	merged.OneOf = typed.OneOf
	merged.AnyOf = typed.AnyOf

	for name, def := range compRoot.Definitions {
		merged.Definitions[name] = def
	}
	for name, def := range typedRoot.Definitions {
		merged.Definitions[name] = def
	}
	delete(merged.Definitions, refDefinitionName(compRoot.Ref))
	delete(merged.Definitions, refDefinitionName(typedRoot.Ref))
	pruneUnreachableDefinitions(merged)

	return merged, nil
}

// checkUnhandledConstraints errors when a reflected root definition carries
// validation keywords the flattening does not merge. Without this, a future
// JSONSchemaExtend hook setting e.g. if/then or dependentRequired would be
// silently dropped from the published schema, loosening it with no failure.
func checkUnhandledConstraints(def *jsonschema.Schema, name string) error {
	unhandled := []struct {
		keyword string
		set     bool
	}{
		{"allOf", len(def.AllOf) > 0},
		{"not", def.Not != nil},
		{"if", def.If != nil},
		{"then", def.Then != nil},
		{"else", def.Else != nil},
		{"dependentSchemas", len(def.DependentSchemas) > 0},
		{"dependentRequired", len(def.DependentRequired) > 0},
		{"patternProperties", len(def.PatternProperties) > 0},
		{"propertyNames", def.PropertyNames != nil},
		{"enum", len(def.Enum) > 0},
		{"const", def.Const != nil},
		{"extras", len(def.Extras) > 0},
	}
	for _, u := range unhandled {
		if u.set {
			return fmt.Errorf("%s declares root-level %q, which flattenedComponentSchema does not merge; extend the flattening before using it", name, u.keyword)
		}
	}
	return nil
}

func rootDefinition(schm *jsonschema.Schema) (*jsonschema.Schema, error) {
	name := refDefinitionName(schm.Ref)
	def, ok := schm.Definitions[name]
	if !ok {
		return nil, fmt.Errorf("reflected schema has no root definition %q", name)
	}
	return def, nil
}

func refDefinitionName(ref string) string {
	return strings.TrimPrefix(ref, "#/$defs/")
}

func mergeRequired(lists ...[]string) []string {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, list := range lists {
		for _, k := range list {
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, k)
		}
	}
	return out
}

var defRefPattern = regexp.MustCompile(`#/\$defs/([^"]+)`)

// pruneUnreachableDefinitions drops $defs entries that are not transitively
// referenced from the schema root, e.g. the per-type config definitions left
// over after the nested component properties are removed.
func pruneUnreachableDefinitions(schm *jsonschema.Schema) {
	root := *schm
	root.Definitions = nil

	reachable := make(map[string]bool)
	queue := refsIn(&root)
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if reachable[name] {
			continue
		}
		reachable[name] = true
		if def, ok := schm.Definitions[name]; ok {
			queue = append(queue, refsIn(def)...)
		}
	}

	for name := range schm.Definitions {
		if !reachable[name] {
			delete(schm.Definitions, name)
		}
	}
}

func refsIn(schm *jsonschema.Schema) []string {
	b, err := json.Marshal(schm)
	if err != nil {
		return nil
	}

	names := make([]string, 0)
	for _, m := range defRefPattern.FindAllStringSubmatch(string(b), -1) {
		names = append(names, m[1])
	}
	return names
}
