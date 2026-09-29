package config

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

type SchemaBuilder struct {
	schema *jsonschema.Schema
}

func NewSchemaBuilder(s *jsonschema.Schema) *SchemaBuilder {
	return &SchemaBuilder{schema: s}
}

func (sb *SchemaBuilder) Field(name string) *FieldBuilder {
	return &FieldBuilder{
		schema: sb.schema,
		name:   name,
	}
}

type FieldBuilder struct {
	schema *jsonschema.Schema
	name   string
}

func (fb *FieldBuilder) getProperty() *jsonschema.Schema {
	prop, ok := fb.schema.Properties.Get(fb.name)
	if !ok {
		prop = &jsonschema.Schema{}
		fb.schema.Properties.Set(fb.name, prop)
	}
	return prop
}

func (fb *FieldBuilder) setProperty(prop *jsonschema.Schema) *FieldBuilder {
	fb.schema.Properties.Set(fb.name, prop)
	return fb
}

func (fb *FieldBuilder) Short(desc string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Description = desc
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Long(desc string) *FieldBuilder {
	prop := fb.getProperty()
	if prop.Description != "" {
		prop.Description = prop.Description + "\n" + desc
	} else {
		prop.Description = desc
	}
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Example(v any) *FieldBuilder {
	prop := fb.getProperty()
	if prop.Examples == nil {
		prop.Examples = []any{}
	}
	prop.Examples = append(prop.Examples, v)
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Deprecated(reason string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Deprecated = true
	if reason != "" {
		if prop.Description != "" {
			prop.Description = prop.Description + " [DEPRECATED: " + reason + "]"
		} else {
			prop.Description = "[DEPRECATED: " + reason + "]"
		}
	}
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Required() *FieldBuilder {
	found := false
	for _, req := range fb.schema.Required {
		if req == fb.name {
			found = true
			break
		}
	}
	if !found {
		fb.schema.Required = append(fb.schema.Required, fb.name)
	}
	return fb
}

func (fb *FieldBuilder) Enum(values ...any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Enum = values
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Format(format string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Format = format
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Pattern(regex string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Pattern = regex
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Type(schemaType string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Type = schemaType
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) MinLength(min uint64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MinLength = &min
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) MaxLength(max uint64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MaxLength = &max
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Minimum(min float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.Minimum = json.Number(fmt.Sprintf("%g", min))
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Maximum(max float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.Maximum = json.Number(fmt.Sprintf("%g", max))
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) ExclusiveMinimum(min float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.ExclusiveMinimum = json.Number(fmt.Sprintf("%g", min))
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) ExclusiveMaximum(max float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.ExclusiveMaximum = json.Number(fmt.Sprintf("%g", max))
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Items(itemSchema *jsonschema.Schema) *FieldBuilder {
	prop := fb.getProperty()
	prop.Items = itemSchema
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) MinItems(min uint64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MinItems = &min
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Object(callback func(*SchemaBuilder)) *FieldBuilder {
	prop := fb.getProperty()
	prop.Type = "object"
	if prop.Properties == nil {
		prop.Properties = jsonschema.NewProperties()
	}
	nestedBuilder := &SchemaBuilder{schema: prop}
	callback(nestedBuilder)
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) ObjectRequired(requiredFields ...string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Required = requiredFields
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Default(value any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Default = value
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Title(title string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Title = title
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Description(desc string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Description = desc
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Field(name string) *FieldBuilder {
	return &FieldBuilder{
		schema: fb.schema,
		name:   name,
	}
}

func (fb *FieldBuilder) SchemaBackref() *SchemaBuilder {
	return &SchemaBuilder{schema: fb.schema}
}

func (fb *FieldBuilder) Const(value any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Const = value
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) MultipleOf(multiple float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MultipleOf = json.Number(fmt.Sprintf("%g", multiple))
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) ReadOnly(readOnly bool) *FieldBuilder {
	prop := fb.getProperty()
	prop.ReadOnly = readOnly
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) WriteOnly(writeOnly bool) *FieldBuilder {
	prop := fb.getProperty()
	prop.WriteOnly = writeOnly
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) Examples(values ...any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Examples = values
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) OneOfRequired(groupName string) *FieldBuilder {
	prop := fb.getProperty()
	if prop.Extras == nil {
		prop.Extras = map[string]interface{}{}
	}
	prop.Extras["oneof_required"] = groupName
	return fb.setProperty(prop)
}

// TODO(fd): how have we avoided such a solution to date? a default value seems preferable.

func (fb *FieldBuilder) Nullable() *FieldBuilder {
	prop := fb.getProperty()
	existing := *prop
	existing.Description = ""
	*prop = jsonschema.Schema{
		Description: prop.Description,
		OneOf: []*jsonschema.Schema{
			&existing,
			{Type: "null"},
		},
	}
	return fb.setProperty(prop)
}

func (fb *FieldBuilder) String() string {
	return fmt.Sprintf("FieldBuilder{name=%q}", fb.name)
}
