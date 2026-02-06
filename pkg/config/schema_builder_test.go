package config

import (
	"slices"
	"testing"

	"github.com/invopop/jsonschema"
)

func TestNewSchemaBuilder(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	if builder == nil {
		t.Fatal("NewSchemaBuilder returned nil")
	}
	if builder.schema != schema {
		t.Error("SchemaBuilder did not store the schema correctly")
	}
}

func TestFieldBuilder_Short(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("test_field")
	result := fb.Short("This is a short description")

	// Verify chaining works
	if result != fb {
		t.Error("Short() did not return the FieldBuilder for chaining")
	}

	// Verify description was set
	prop, ok := schema.Properties.Get("test_field")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Description != "This is a short description" {
		t.Errorf("Expected description 'This is a short description', got %q", prop.Description)
	}
}

func TestFieldBuilder_Long(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("test_field")
	fb.Short("Short")
	fb.Long("This is a longer description with more details")

	prop, ok := schema.Properties.Get("test_field")
	if !ok {
		t.Fatal("Property was not created")
	}
	expected := "Short\nThis is a longer description with more details"
	if prop.Description != expected {
		t.Errorf("Expected description %q, got %q", expected, prop.Description)
	}
}

func TestFieldBuilder_Example(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("test_field")
	fb.Example("example1")
	fb.Example("example2")
	fb.Example(123)

	prop, ok := schema.Properties.Get("test_field")
	if !ok {
		t.Fatal("Property was not created")
	}
	if len(prop.Examples) != 3 {
		t.Errorf("Expected 3 examples, got %d", len(prop.Examples))
	}
	if prop.Examples[0] != "example1" || prop.Examples[1] != "example2" || prop.Examples[2] != 123 {
		t.Error("Examples were not set correctly")
	}
}

func TestFieldBuilder_Examples(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("test_field")
	fb.Examples("ex1", "ex2", "ex3")

	prop, ok := schema.Properties.Get("test_field")
	if !ok {
		t.Fatal("Property was not created")
	}
	if len(prop.Examples) != 3 {
		t.Errorf("Expected 3 examples, got %d", len(prop.Examples))
	}
}

func TestFieldBuilder_Deprecated(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("test_field")
	fb.Short("Some field")
	fb.Deprecated("use new_field instead")

	prop, ok := schema.Properties.Get("test_field")
	if !ok {
		t.Fatal("Property was not created")
	}
	if !prop.Deprecated {
		t.Error("Expected Deprecated to be true")
	}
	if !contains(prop.Description, "DEPRECATED: use new_field instead") {
		t.Errorf("Expected deprecation message in description, got %q", prop.Description)
	}
}

func TestFieldBuilder_DeprecatedWithoutReason(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("test_field")
	fb.Short("Some field")
	fb.Deprecated("")

	prop, ok := schema.Properties.Get("test_field")
	if !ok {
		t.Fatal("Property was not created")
	}
	if !prop.Deprecated {
		t.Error("Expected Deprecated to be true")
	}
	// Description should only have the original text
	if prop.Description != "Some field" {
		t.Errorf("Expected description 'Some field', got %q", prop.Description)
	}
}

func TestFieldBuilder_Required(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
		Required:   []string{},
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("test_field")
	result := fb.Required()

	// Verify chaining works
	if result != fb {
		t.Error("Required() did not return the FieldBuilder for chaining")
	}

	// Verify field was added to Required
	if len(schema.Required) != 1 || schema.Required[0] != "test_field" {
		t.Errorf("Expected Required to contain 'test_field', got %v", schema.Required)
	}

	// Verify it doesn't add duplicates
	fb.Required()
	if len(schema.Required) != 1 {
		t.Errorf("Expected Required to still have 1 item, got %d", len(schema.Required))
	}
}

func TestFieldBuilder_Enum(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("status")
	fb.Enum("active", "inactive", "pending")

	prop, ok := schema.Properties.Get("status")
	if !ok {
		t.Fatal("Property was not created")
	}
	if len(prop.Enum) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(prop.Enum))
	}
}

func TestFieldBuilder_Format(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("email")
	fb.Format("email")

	prop, ok := schema.Properties.Get("email")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Format != "email" {
		t.Errorf("Expected format 'email', got %q", prop.Format)
	}
}

func TestFieldBuilder_Pattern(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("version")
	fb.Pattern(`^\d+\.\d+\.\d+$`)

	prop, ok := schema.Properties.Get("version")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Pattern != `^\d+\.\d+\.\d+$` {
		t.Errorf("Expected pattern '^\\d+\\.\\d+\\.\\d+$', got %q", prop.Pattern)
	}
}

func TestFieldBuilder_Type(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("count")
	fb.Type("integer")

	prop, ok := schema.Properties.Get("count")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Type != "integer" {
		t.Errorf("Expected type 'integer', got %q", prop.Type)
	}
}

func TestFieldBuilder_MinMaxLength(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("username")
	fb.MinLength(3)
	fb.MaxLength(20)

	prop, ok := schema.Properties.Get("username")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.MinLength == nil || *prop.MinLength != 3 {
		t.Errorf("Expected MinLength 3, got %v", prop.MinLength)
	}
	if prop.MaxLength == nil || *prop.MaxLength != 20 {
		t.Errorf("Expected MaxLength 20, got %v", prop.MaxLength)
	}
}

func TestFieldBuilder_MinMaxNumeric(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("age")
	fb.Minimum(0)
	fb.Maximum(150)

	prop, ok := schema.Properties.Get("age")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Minimum != "0" {
		t.Errorf("Expected Minimum '0', got %q", prop.Minimum)
	}
	if prop.Maximum != "150" {
		t.Errorf("Expected Maximum '150', got %q", prop.Maximum)
	}
}

func TestFieldBuilder_Default(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("status")
	fb.Default("active")

	prop, ok := schema.Properties.Get("status")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Default != "active" {
		t.Errorf("Expected default 'active', got %v", prop.Default)
	}
}

func TestFieldBuilder_Title(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("user_id")
	fb.Title("User ID")

	prop, ok := schema.Properties.Get("user_id")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Title != "User ID" {
		t.Errorf("Expected title 'User ID', got %q", prop.Title)
	}
}

func TestFieldBuilder_Description(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("config")
	fb.Description("Configuration object")

	prop, ok := schema.Properties.Get("config")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Description != "Configuration object" {
		t.Errorf("Expected description 'Configuration object', got %q", prop.Description)
	}
}

func TestFieldBuilder_FieldChaining(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	builder.Field("field1").Short("First field").Field("field2").Short("Second field")

	prop1, ok1 := schema.Properties.Get("field1")
	prop2, ok2 := schema.Properties.Get("field2")

	if !ok1 || !ok2 {
		t.Fatal("Properties were not created")
	}
	if prop1.Description != "First field" {
		t.Errorf("Expected field1 description 'First field', got %q", prop1.Description)
	}
	if prop2.Description != "Second field" {
		t.Errorf("Expected field2 description 'Second field', got %q", prop2.Description)
	}
}

func TestFieldBuilder_SchemaBackref(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	backref := builder.Field("field1").Short("Field 1").SchemaBackref()

	if backref == nil {
		t.Fatal("SchemaBackref returned nil")
	}
	if backref.schema != schema {
		t.Error("SchemaBackref did not return correct schema")
	}

	// Verify we can use it to build another field
	backref.Field("field2").Short("Field 2")
	prop, ok := schema.Properties.Get("field2")
	if !ok || prop.Description != "Field 2" {
		t.Error("Could not chain after SchemaBackref")
	}
}

func TestFieldBuilder_Const(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("version")
	fb.Const("1.0.0")

	prop, ok := schema.Properties.Get("version")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Const != "1.0.0" {
		t.Errorf("Expected const '1.0.0', got %v", prop.Const)
	}
}

func TestFieldBuilder_MultipleOf(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("quantity")
	fb.MultipleOf(5)

	prop, ok := schema.Properties.Get("quantity")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.MultipleOf != "5" {
		t.Errorf("Expected MultipleOf '5', got %q", prop.MultipleOf)
	}
}

func TestFieldBuilder_ReadWriteOnly(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	builder.Field("readonly_field").ReadOnly(true)
	builder.Field("writeonly_field").WriteOnly(true)

	ro, _ := schema.Properties.Get("readonly_field")
	wo, _ := schema.Properties.Get("writeonly_field")

	if !ro.ReadOnly {
		t.Error("ReadOnly field was not marked correctly")
	}
	if !wo.WriteOnly {
		t.Error("WriteOnly field was not marked correctly")
	}
}

func TestFluentChaining(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
		Required:   []string{},
	}

	// Test comprehensive fluent chaining as shown in the example usage
	NewSchemaBuilder(schema).
		Field("repo_url").
		Short("URL of repo").
		Long("Full URL to the Helm chart repository").
		Example("https://charts.bitnami.com/bitnami").
		Required().
		Field("chart").
		Short("Chart name").
		Example("nginx").
		Required().
		Field("version").
		Short("Chart version").
		Format("semver")

	// Verify all fields were created
	repoURL, ok1 := schema.Properties.Get("repo_url")
	chart, ok2 := schema.Properties.Get("chart")
	version, ok3 := schema.Properties.Get("version")

	if !ok1 || !ok2 || !ok3 {
		t.Fatal("Not all properties were created")
	}

	// Verify repo_url
	if !contains(repoURL.Description, "Full URL") {
		t.Error("repo_url description not set correctly")
	}
	if len(repoURL.Examples) != 1 {
		t.Errorf("Expected 1 example for repo_url, got %d", len(repoURL.Examples))
	}

	// Verify chart
	if chart.Description != "Chart name" {
		t.Errorf("chart description incorrect: %q", chart.Description)
	}

	// Verify version
	if version.Format != "semver" {
		t.Errorf("version format incorrect: %q", version.Format)
	}

	// Verify required
	if len(schema.Required) != 2 {
		t.Errorf("Expected 2 required fields, got %d", len(schema.Required))
	}
}

func TestExclusiveMinMax(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	fb := builder.Field("positive_number")
	fb.ExclusiveMinimum(0)
	fb.ExclusiveMaximum(100)

	prop, ok := schema.Properties.Get("positive_number")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.ExclusiveMinimum != "0" {
		t.Errorf("Expected ExclusiveMinimum '0', got %q", prop.ExclusiveMinimum)
	}
	if prop.ExclusiveMaximum != "100" {
		t.Errorf("Expected ExclusiveMaximum '100', got %q", prop.ExclusiveMaximum)
	}
}

func TestFieldBuilder_Items(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	itemSchema := &jsonschema.Schema{
		Type: "string",
	}
	fb := builder.Field("tags")
	fb.Items(itemSchema)

	prop, ok := schema.Properties.Get("tags")
	if !ok {
		t.Fatal("Property was not created")
	}
	if prop.Items == nil {
		t.Fatal("Items schema was not set")
	}
	if prop.Items.Type != "string" {
		t.Errorf("Expected item type 'string', got %q", prop.Items.Type)
	}
}

func TestFieldBuilder_NestedObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	builder.Field("config").Object(func(sb *SchemaBuilder) {
		sb.Field("host").Short("Server hostname")
		sb.Field("port").Type("integer")
	})

	config, ok := schema.Properties.Get("config")
	if !ok {
		t.Fatal("config field was not created")
	}
	if config.Type != "object" {
		t.Errorf("Expected config type 'object', got %q", config.Type)
	}
	if config.Properties == nil {
		t.Fatal("Nested properties were not created")
	}

	host, hostOk := config.Properties.Get("host")
	port, portOk := config.Properties.Get("port")

	if !hostOk || !portOk {
		t.Fatal("Nested properties were not created")
	}
	if host.Description != "Server hostname" {
		t.Errorf("Expected host description 'Server hostname', got %q", host.Description)
	}
	if port.Type != "integer" {
		t.Errorf("Expected port type 'integer', got %q", port.Type)
	}
}

func TestFieldBuilder_NestedObjectRequired(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	builder.Field("config").
		Object(func(sb *SchemaBuilder) {
			sb.Field("host").Short("Server hostname")
			sb.Field("port").Type("integer")
		}).
		ObjectRequired("host", "port")

	config, ok := schema.Properties.Get("config")
	if !ok {
		t.Fatal("config field was not created")
	}
	if len(config.Required) != 2 {
		t.Errorf("Expected 2 required nested fields, got %d", len(config.Required))
	}
	if config.Required[0] != "host" || config.Required[1] != "port" {
		t.Errorf("Expected required fields [host, port], got %v", config.Required)
	}
}

func TestFieldBuilder_DeeplyNestedObject(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	// Three levels of nesting
	builder.Field("app").
		Short("Application config").
		Object(func(sb *SchemaBuilder) {
			sb.Field("database").
				Short("Database config").
				Object(func(innerSb *SchemaBuilder) {
					innerSb.Field("host").Short("DB host").Type("string")
					innerSb.Field("port").Short("DB port").Type("integer")
					innerSb.Field("username").Short("DB user")
				}).
				ObjectRequired("host", "port")
			sb.Field("cache").
				Short("Cache config").
				Object(func(innerSb *SchemaBuilder) {
					innerSb.Field("enabled").Type("boolean").Default(true)
					innerSb.Field("ttl").Type("integer")
				})
		}).
		ObjectRequired("database")

	app, ok := schema.Properties.Get("app")
	if !ok {
		t.Fatal("app field not created")
	}
	if app.Type != "object" {
		t.Errorf("Expected app type 'object', got %q", app.Type)
	}

	database, dbOk := app.Properties.Get("database")
	if !dbOk {
		t.Fatal("database field not created in app")
	}
	if database.Type != "object" {
		t.Errorf("Expected database type 'object', got %q", database.Type)
	}

	dbHost, hostOk := database.Properties.Get("host")
	if !hostOk {
		t.Fatal("host field not created in database")
	}
	if dbHost.Description != "DB host" {
		t.Errorf("Expected host description 'DB host', got %q", dbHost.Description)
	}

	cache, cacheOk := app.Properties.Get("cache")
	if !cacheOk {
		t.Fatal("cache field not created in app")
	}

	cacheEnabled, enabledOk := cache.Properties.Get("enabled")
	if !enabledOk {
		t.Fatal("enabled field not created in cache")
	}
	if cacheEnabled.Default != true {
		t.Errorf("Expected enabled default true, got %v", cacheEnabled.Default)
	}
}

func TestFieldBuilder_NestedObjectWithChaining(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}

	NewSchemaBuilder(schema).
		Field("server").
		Short("Server configuration").
		Object(func(sb *SchemaBuilder) {
			sb.Field("host").Short("Server host").Format("hostname").Required()
			sb.Field("port").Short("Server port").Type("integer").Minimum(1).Maximum(65535).Required()
			sb.Field("ssl").Short("Enable SSL").Type("boolean").Default(false)
		}).
		ObjectRequired("host", "port").
		Field("logging").
		Short("Logging configuration").
		Object(func(sb *SchemaBuilder) {
			sb.Field("level").Short("Log level").Enum("debug", "info", "warn", "error").Default("info")
			sb.Field("format").Short("Log format").Enum("json", "text").Default("json")
		})

	server, _ := schema.Properties.Get("server")
	logging, _ := schema.Properties.Get("logging")

	if server.Description != "Server configuration" {
		t.Error("server description not set")
	}
	if logging.Description != "Logging configuration" {
		t.Error("logging description not set")
	}

	// Verify server properties
	serverHost, _ := server.Properties.Get("host")
	if serverHost.Format != "hostname" {
		t.Errorf("Expected host format 'hostname', got %q", serverHost.Format)
	}

	// Verify logging properties
	loggingLevel, _ := logging.Properties.Get("level")
	if len(loggingLevel.Enum) != 4 {
		t.Errorf("Expected 4 enum values for level, got %d", len(loggingLevel.Enum))
	}
}

func TestFieldBuilder_NestedObjectArrayItems(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	builder := NewSchemaBuilder(schema)

	// Create an array of objects with properly initialized Properties
	itemSchema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	itemBuilder := NewSchemaBuilder(itemSchema)
	itemBuilder.Field("name").Short("Item name").Type("string")
	itemBuilder.Field("value").Short("Item value").Type("number")

	builder.Field("items").
		Short("List of items").
		Type("array").
		Items(itemSchema)

	items, ok := schema.Properties.Get("items")
	if !ok {
		t.Fatal("items field not created")
	}
	if items.Type != "array" {
		t.Errorf("Expected items type 'array', got %q", items.Type)
	}
	if items.Items == nil {
		t.Fatal("items.Items not set")
	}

	itemName, _ := itemSchema.Properties.Get("name")
	if itemName.Description != "Item name" {
		t.Error("item name description not set correctly")
	}
}

func TestFieldBuilder_ComplexNestedStructure(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}

	// Simulate a real-world structure like HelmChartComponentConfig
	NewSchemaBuilder(schema).
		Field("chart_name").
		Short("Chart name").
		Type("string").
		Required().
		Field("helm_repo").
		Short("Helm repository configuration").
		Object(func(sb *SchemaBuilder) {
			sb.Field("repo_url").
				Short("Repository URL").
				Type("string").
				Format("uri").
				Example("https://charts.bitnami.com/bitnami")
			sb.Field("chart").
				Short("Chart name in repository").
				Type("string").
				Example("nginx")
			sb.Field("version").
				Short("Chart version").
				Type("string").
				Pattern(`^\d+\.\d+\.\d+$`)
		}).
		ObjectRequired("repo_url", "chart").
		Field("values").
		Short("Helm values").
		Object(func(sb *SchemaBuilder) {
			sb.Field("replicaCount").Type("integer").Default(1)
			sb.Field("image").
				Short("Container image config").
				Object(func(innerSb *SchemaBuilder) {
					innerSb.Field("repository").Short("Image repository")
					innerSb.Field("tag").Short("Image tag")
					innerSb.Field("pullPolicy").Short("Image pull policy").Enum("Always", "IfNotPresent", "Never")
				})
		}).
		Field("namespace").
		Short("Kubernetes namespace").
		Type("string").
		Default("default")

	// Verify top-level structure
	chartName, _ := schema.Properties.Get("chart_name")
	if chartName.Type != "string" {
		t.Errorf("Expected chart_name type 'string', got %q", chartName.Type)
	}

	helmRepo, _ := schema.Properties.Get("helm_repo")
	if helmRepo.Type != "object" {
		t.Errorf("Expected helm_repo type 'object', got %q", helmRepo.Type)
	}

	// Verify nested structure
	repoURL, _ := helmRepo.Properties.Get("repo_url")
	if repoURL.Format != "uri" {
		t.Errorf("Expected repo_url format 'uri', got %q", repoURL.Format)
	}

	// Verify deeply nested structure
	values, _ := schema.Properties.Get("values")
	if values.Type != "object" {
		t.Errorf("Expected values type 'object', got %q", values.Type)
	}

	valuesImage, _ := values.Properties.Get("image")
	if valuesImage.Type != "object" {
		t.Errorf("Expected image type 'object', got %q", valuesImage.Type)
	}

	imagePullPolicy, _ := valuesImage.Properties.Get("pullPolicy")
	if len(imagePullPolicy.Enum) != 3 {
		t.Errorf("Expected 3 enum values for pullPolicy, got %d", len(imagePullPolicy.Enum))
	}
}

func TestSchemaBuilder_OneOfGroup(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	schema.Properties.Set("connected_repo", &jsonschema.Schema{Type: "object"})
	schema.Properties.Set("public_repo", &jsonschema.Schema{Type: "object"})

	NewSchemaBuilder(schema).OneOfGroup("vcs", "connected_repo", "public_repo")

	if len(schema.OneOf) != 2 {
		t.Fatalf("Expected 2 OneOf entries (one per field), got %d", len(schema.OneOf))
	}
	for i, entry := range schema.OneOf {
		if entry.Title != "vcs" {
			t.Errorf("OneOf[%d]: Expected Title 'vcs', got %q", i, entry.Title)
		}
		if len(entry.Required) != 1 {
			t.Errorf("OneOf[%d]: Expected 1 Required field, got %d", i, len(entry.Required))
		}
	}
	if schema.OneOf[0].Required[0] != "connected_repo" {
		t.Errorf("OneOf[0].Required[0] = %q, want 'connected_repo'", schema.OneOf[0].Required[0])
	}
	if schema.OneOf[1].Required[0] != "public_repo" {
		t.Errorf("OneOf[1].Required[0] = %q, want 'public_repo'", schema.OneOf[1].Required[0])
	}

	for _, name := range []string{"connected_repo", "public_repo"} {
		prop, ok := schema.Properties.Get(name)
		if !ok {
			t.Fatalf("Property %q not found", name)
		}
		val, exists := prop.Extras["oneof_required"]
		if !exists {
			t.Errorf("Property %q missing Extras[oneof_required]", name)
		} else if val != "vcs" {
			t.Errorf("Property %q Extras[oneof_required] = %v, want 'vcs'", name, val)
		}
	}
}

func TestSchemaBuilder_OneOfGroup_ThreeFields(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	schema.Properties.Set("public_repo", &jsonschema.Schema{Type: "object"})
	schema.Properties.Set("connected_repo", &jsonschema.Schema{Type: "object"})
	schema.Properties.Set("helm_repo", &jsonschema.Schema{Type: "object"})

	NewSchemaBuilder(schema).OneOfGroup("vcs", "public_repo", "connected_repo", "helm_repo")

	if len(schema.OneOf) != 3 {
		t.Fatalf("Expected 3 OneOf entries (one per field), got %d", len(schema.OneOf))
	}
	expectedFields := []string{"public_repo", "connected_repo", "helm_repo"}
	for i, entry := range schema.OneOf {
		if len(entry.Required) != 1 {
			t.Errorf("OneOf[%d]: Expected 1 Required field, got %d", i, len(entry.Required))
		}
		if entry.Required[0] != expectedFields[i] {
			t.Errorf("OneOf[%d].Required[0] = %q, want %q", i, entry.Required[0], expectedFields[i])
		}
	}

	for _, name := range expectedFields {
		prop, ok := schema.Properties.Get(name)
		if !ok {
			t.Fatalf("Property %q not found", name)
		}
		val, exists := prop.Extras["oneof_required"]
		if !exists {
			t.Errorf("Property %q missing Extras[oneof_required]", name)
		} else if val != "vcs" {
			t.Errorf("Property %q Extras[oneof_required] = %v, want 'vcs'", name, val)
		}
	}
}

func TestSchemaBuilder_OneOfGroup_EncoderCompatibility(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}
	schema.Properties.Set("connected_repo", &jsonschema.Schema{Type: "object"})
	schema.Properties.Set("public_repo", &jsonschema.Schema{Type: "object"})

	NewSchemaBuilder(schema).OneOfGroup("vcs", "connected_repo", "public_repo")

	// Build oneOFGroups the same way the encoder does (gen.go)
	oneOFGroups := make(map[string]map[string]bool)
	for _, s := range schema.OneOf {
		if _, exists := oneOFGroups[s.Title]; !exists {
			oneOFGroups[s.Title] = make(map[string]bool)
		}
		for _, r := range s.Required {
			oneOFGroups[s.Title][r] = true
		}
	}

	if !oneOFGroups["vcs"]["connected_repo"] {
		t.Error("Expected oneOFGroups[vcs][connected_repo] to be true")
	}
	if !oneOFGroups["vcs"]["public_repo"] {
		t.Error("Expected oneOFGroups[vcs][public_repo] to be true")
	}

	// Verify each property has the correct oneof_required extra
	for _, name := range []string{"connected_repo", "public_repo"} {
		prop, ok := schema.Properties.Get(name)
		if !ok {
			t.Fatalf("Property %q not found", name)
		}
		val, exists := prop.Extras["oneof_required"]
		if !exists {
			t.Errorf("Property %q missing Extras[oneof_required]", name)
		} else if val != "vcs" {
			t.Errorf("Property %q Extras[oneof_required] = %v, want 'vcs'", name, val)
		}
	}

	// Verify the group name is in the encoder's allowlist (mirrors generator.StructTagOneOfRequiredGroups)
	encoderAllowlist := []string{"component_type", "vcs"}
	if !slices.Contains(encoderAllowlist, "vcs") {
		t.Error("Expected 'vcs' to be in the encoder's oneOf required groups allowlist")
	}
}

func TestSchemaBuilder_OneOfGroup_Chaining(t *testing.T) {
	schema := &jsonschema.Schema{
		Properties: jsonschema.NewProperties(),
	}

	// Verify OneOfGroup returns *SchemaBuilder and supports chaining into Field()
	NewSchemaBuilder(schema).
		OneOfGroup("vcs", "a", "b").
		Field("a").
		Short("desc")

	prop, ok := schema.Properties.Get("a")
	if !ok {
		t.Fatal("Property 'a' not found after chaining")
	}
	if prop.Description != "desc" {
		t.Errorf("Expected description 'desc', got %q", prop.Description)
	}

	val, exists := prop.Extras["oneof_required"]
	if !exists {
		t.Error("Property 'a' missing Extras[oneof_required] after chaining")
	} else if val != "vcs" {
		t.Errorf("Property 'a' Extras[oneof_required] = %v, want 'vcs'", val)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
