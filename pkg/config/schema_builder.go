package config

import (
	"encoding/json"
	"fmt"

	"github.com/invopop/jsonschema"
)

// SchemaBuilder provides a Cobra-style fluent builder API for constructing JSON Schema properties.
// It wraps a jsonschema.Schema and provides Field() method to start building individual fields.
//
// Example usage with nested objects:
//
//	func (h HelmRepoConfig) JSONSchemaExtend(s *jsonschema.Schema) {
//		NewSchemaBuilder(s).
//			Field("repo_url").
//				Short("URL of repo").
//				Long("Full URL to the Helm chart repository").
//				Example("https://charts.bitnami.com/bitnami").
//				Required().
//			Field("helm_repo").
//				Short("Helm repository configuration").
//				Object(func(sb *SchemaBuilder) {
//					sb.Field("repo_url").Short("Repository URL").Format("uri")
//					sb.Field("chart").Short("Chart name")
//					sb.Field("version").Short("Chart version")
//				}).
//				ObjectRequired("repo_url", "chart")
//	}
type SchemaBuilder struct {
	schema *jsonschema.Schema
}

// NewSchemaBuilder creates a new SchemaBuilder for the given JSON schema.
func NewSchemaBuilder(s *jsonschema.Schema) *SchemaBuilder {
	return &SchemaBuilder{schema: s}
}

// Field returns a FieldBuilder for the property with the given name.
// If the property doesn't exist in the schema, it will be created.
func (sb *SchemaBuilder) Field(name string) *FieldBuilder {
	return &FieldBuilder{
		schema: sb.schema,
		name:   name,
	}
}

// FieldBuilder provides a fluent API for building individual schema field properties.
type FieldBuilder struct {
	schema *jsonschema.Schema
	name   string
}

// getProperty retrieves or creates the property from the schema.
func (fb *FieldBuilder) getProperty() *jsonschema.Schema {
	prop, ok := fb.schema.Properties.Get(fb.name)
	if !ok {
		prop = &jsonschema.Schema{}
		fb.schema.Properties.Set(fb.name, prop)
	}
	return prop
}

// setProperty updates the property in the schema.
func (fb *FieldBuilder) setProperty(prop *jsonschema.Schema) *FieldBuilder {
	fb.schema.Properties.Set(fb.name, prop)
	return fb
}

// Short sets a brief description for the field.
func (fb *FieldBuilder) Short(desc string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Description = desc
	return fb.setProperty(prop)
}

// Long sets a longer description for the field, appending to existing description if present.
func (fb *FieldBuilder) Long(desc string) *FieldBuilder {
	prop := fb.getProperty()
	if prop.Description != "" {
		prop.Description = prop.Description + "\n" + desc
	} else {
		prop.Description = desc
	}
	return fb.setProperty(prop)
}

// Example adds an example value to the field's examples list.
func (fb *FieldBuilder) Example(v any) *FieldBuilder {
	prop := fb.getProperty()
	if prop.Examples == nil {
		prop.Examples = []any{}
	}
	prop.Examples = append(prop.Examples, v)
	return fb.setProperty(prop)
}

// Deprecated marks the field as deprecated with an optional reason message.
// The reason will be appended to the description if provided.
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

// Required marks the field as required in the parent schema.
func (fb *FieldBuilder) Required() *FieldBuilder {
	// Add the field name to the schema's required list if not already present
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

// Enum sets the allowed values for the field.
func (fb *FieldBuilder) Enum(values ...any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Enum = values
	return fb.setProperty(prop)
}

// Format sets the format constraint for the field (e.g., "date-time", "email", "uri").
func (fb *FieldBuilder) Format(format string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Format = format
	return fb.setProperty(prop)
}

// Pattern sets a regex pattern constraint for the field.
func (fb *FieldBuilder) Pattern(regex string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Pattern = regex
	return fb.setProperty(prop)
}

// Type sets the JSON Schema type for the field (e.g., "string", "number", "boolean", "array", "object").
func (fb *FieldBuilder) Type(schemaType string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Type = schemaType
	return fb.setProperty(prop)
}

// MinLength sets the minimum length constraint for string fields.
func (fb *FieldBuilder) MinLength(min uint64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MinLength = &min
	return fb.setProperty(prop)
}

// MaxLength sets the maximum length constraint for string fields.
func (fb *FieldBuilder) MaxLength(max uint64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MaxLength = &max
	return fb.setProperty(prop)
}

// Minimum sets the minimum value constraint for numeric fields.
func (fb *FieldBuilder) Minimum(min float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.Minimum = json.Number(fmt.Sprintf("%g", min))
	return fb.setProperty(prop)
}

// Maximum sets the maximum value constraint for numeric fields.
func (fb *FieldBuilder) Maximum(max float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.Maximum = json.Number(fmt.Sprintf("%g", max))
	return fb.setProperty(prop)
}

// ExclusiveMinimum sets the exclusive minimum value constraint for numeric fields.
func (fb *FieldBuilder) ExclusiveMinimum(min float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.ExclusiveMinimum = json.Number(fmt.Sprintf("%g", min))
	return fb.setProperty(prop)
}

// ExclusiveMaximum sets the exclusive maximum value constraint for numeric fields.
func (fb *FieldBuilder) ExclusiveMaximum(max float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.ExclusiveMaximum = json.Number(fmt.Sprintf("%g", max))
	return fb.setProperty(prop)
}

// Items sets the schema for array items.
func (fb *FieldBuilder) Items(itemSchema *jsonschema.Schema) *FieldBuilder {
	prop := fb.getProperty()
	prop.Items = itemSchema
	return fb.setProperty(prop)
}

// MinItems sets the minimum number of items constraint for array fields.
func (fb *FieldBuilder) MinItems(min uint64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MinItems = &min
	return fb.setProperty(prop)
}

// Object marks the field as an object type and provides a builder for its nested properties.
// The callback function receives a SchemaBuilder for the nested object's properties.
// Example:
//
//	NewSchemaBuilder(s).Field("config").Object(func(sb *SchemaBuilder) {
//		sb.Field("host").Short("Server hostname").Required()
//		sb.Field("port").Short("Server port").Type("integer").Required()
//	})
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

// ObjectRequired marks the field as an object and sets which nested properties are required.
// Use with Object() to set both the nested properties and required fields.
func (fb *FieldBuilder) ObjectRequired(requiredFields ...string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Required = requiredFields
	return fb.setProperty(prop)
}

// Default sets the default value for the field.
func (fb *FieldBuilder) Default(value any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Default = value
	return fb.setProperty(prop)
}

// Title sets the title/display name for the field.
func (fb *FieldBuilder) Title(title string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Title = title
	return fb.setProperty(prop)
}

// Description sets the description for the field (replaces existing description).
func (fb *FieldBuilder) Description(desc string) *FieldBuilder {
	prop := fb.getProperty()
	prop.Description = desc
	return fb.setProperty(prop)
}

// Field returns a new FieldBuilder for a different field within the same schema.
// This allows chaining multiple field definitions.
func (fb *FieldBuilder) Field(name string) *FieldBuilder {
	return &FieldBuilder{
		schema: fb.schema,
		name:   name,
	}
}

// SchemaBackref returns the underlying SchemaBuilder to allow switching back to schema-level operations.
func (fb *FieldBuilder) SchemaBackref() *SchemaBuilder {
	return &SchemaBuilder{schema: fb.schema}
}

// Const sets a constant value constraint for the field.
func (fb *FieldBuilder) Const(value any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Const = value
	return fb.setProperty(prop)
}

// MultipleOf sets the multipleOf constraint for numeric fields.
func (fb *FieldBuilder) MultipleOf(multiple float64) *FieldBuilder {
	prop := fb.getProperty()
	prop.MultipleOf = json.Number(fmt.Sprintf("%g", multiple))
	return fb.setProperty(prop)
}

// ReadOnly marks the field as read-only.
func (fb *FieldBuilder) ReadOnly(readOnly bool) *FieldBuilder {
	prop := fb.getProperty()
	prop.ReadOnly = readOnly
	return fb.setProperty(prop)
}

// WriteOnly marks the field as write-only.
func (fb *FieldBuilder) WriteOnly(writeOnly bool) *FieldBuilder {
	prop := fb.getProperty()
	prop.WriteOnly = writeOnly
	return fb.setProperty(prop)
}

// Examples sets the complete examples list for the field (replacing any existing examples).
func (fb *FieldBuilder) Examples(values ...any) *FieldBuilder {
	prop := fb.getProperty()
	prop.Examples = values
	return fb.setProperty(prop)
}

// OneOfRequired marks the field as part of a one-of constraint group with the given group name.
// This is used to indicate that exactly one field in the group must be present.
// The group name should match the value in the struct tag (e.g., jsonschema:"oneof_required=group_name").
func (fb *FieldBuilder) OneOfRequired(groupName string) *FieldBuilder {
	prop := fb.getProperty()
	if prop.Extras == nil {
		prop.Extras = map[string]interface{}{}
	}
	prop.Extras["oneof_required"] = groupName
	return fb.setProperty(prop)
}

// String returns a string representation of the field builder for debugging.
func (fb *FieldBuilder) String() string {
	return fmt.Sprintf("FieldBuilder{name=%q}", fb.name)
}
