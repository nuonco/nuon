# Usage Guide

## Quick Start

### 1. Using go generate (Recommended)

The simplest way to regenerate documentation:

```bash
# From the repository root
go generate ./pkg/config/schema
```

This will:
- Read all schemas from `pkg/config/schema/types.go`
- Generate markdown documentation in `docs/config-ref/`
- Use the Mintlify format by default

### 2. Direct Execution

For more control, run the generator directly:

```bash
# Generate with default settings
go run ./pkg/gen/markdown-docs

# Specify output directory
go run ./pkg/gen/markdown-docs -output=custom/output/path

# Use different format
go run ./pkg/gen/markdown-docs -format=github

# Enable verbose logging
go run ./pkg/gen/markdown-docs -verbose
```

## Schema Resolution

The generator uses the **exact same** schema lookup as the API endpoint:

```go
// Both use this method from pkg/config/schema/types.go
schm, err := schema.LookupSchemaType(typ)
```

This means:
- ✅ Documentation always matches API responses
- ✅ Changes to schemas automatically propagate
- ✅ No duplication of schema definitions

## How It Works

### 1. Schema Discovery

The generator reads `schema.SchemaMapping` which contains all config types:

```go
var SchemaMapping = map[string]func() (*jsonschema.Schema, error){
    "action":              ActionConfigSchema,
    "helm":                HelmConfigSchema,
    "terraform":           TerraformModuleConfigSchema,
    // ... 24 total schemas
}
```

### 2. Schema Resolution

For each schema type, the generator:

1. Calls `schema.LookupSchemaType(name)` - same as API
2. Resolves `$ref` definitions to find actual properties
3. Iterates through properties using `Properties.Oldest()` (ordered map)

### 3. Markdown Generation

Uses the markdown AST to build structured documents:

```go
doc := mdast.NewDocument()
doc.AddFrontmatter(map[string]string{
    "title": "Action",
    "description": "JSON Schema reference for action configuration",
})
doc.AddHeading(1, "Action")
doc.AddTable(propertiesTable)
```

### 4. Output

Writes two types of files:

- **Individual schemas**: `action.mdx`, `helm.mdx`, etc.
- **Index page**: `index.mdx` with navigation

## Example Output

### Properties Table

The generator creates clean markdown tables:

```markdown
| Property | Type | Required | Description | Default | Example |
|----------|------|----------|-------------|---------|--------|
| `name` | `string` | ✅ Yes | Action name | - | `"healthcheck"` |
| `timeout` | `string` | ✅ Yes | Max execution time | - | `"5m"` |
```

### Property Details

For enums and multiple examples:

```markdown
### Property Details

#### `type`

**Allowed values:**

- `"provision"`
- `"maintenance"`
- `"deprovision"`
```

## Integration with Workflow

### Recommended Workflow

1. **Make schema changes** in `pkg/config/`
2. **Run codegen**: `./run-nuonctl.sh scripts reset-generated-code`
3. **Generate docs**: `go generate ./pkg/config/schema`
4. **Verify output**: Check `docs/config-ref/` files
5. **Commit all changes**: Schema + generated docs

### CI Integration

Add a check to ensure docs are up-to-date:

```yaml
# .github/workflows/check-docs.yml
- name: Generate docs
  run: go generate ./pkg/config/schema

- name: Check for uncommitted changes
  run: |
    git diff --exit-code docs/config-ref/ || \
      (echo "Documentation is out of date. Run: go generate ./pkg/config/schema" && exit 1)
```

## Customization

### Custom Output Format

To add a new format, extend the generator:

```go
// In main.go
if format == "custom" {
    // Your custom rendering logic
}
```

### Custom AST Nodes

Add new node types to `mdast/ast.go`:

```go
type Alert struct {
    Type string  // "info", "warning", "error"
    Text string
}

func (a *Alert) Render() string {
    return fmt.Sprintf("<Alert type=\"%s\">%s</Alert>\n\n", a.Type, a.Text)
}
```

### Schema-Specific Customization

Add special handling for specific schemas:

```go
// In generateSchemaDoc()
if schemaName == "action" {
    doc.AddParagraph("Actions allow you to run custom scripts...")
}
```

## Troubleshooting

### Documentation Not Updating

1. Check that `go generate` is running successfully
2. Verify output directory exists and is writable
3. Run with `-verbose` flag to see detailed logging

### Schema Not Found

If a schema is missing:

1. Check it's in `schema.SchemaMapping`
2. Verify the schema function returns no error
3. Test the schema endpoint: `GET /v1/general/config-schema?type=<name>`

### MDX Parsing Errors

If Mintlify shows parsing errors:

1. Check that special characters are escaped (handled automatically)
2. Verify table formatting is correct
3. Look for unescaped `{`, `}`, `<`, `>` characters

### Format Issues

For formatting problems:

1. Run `go fmt ./pkg/gen/markdown-docs/...`
2. Check table alignment (should be automatic)
3. Verify newlines between sections

## Performance

The generator is fast:

- **24 schemas**: < 1 second
- **Output**: ~1-5KB per schema file
- **Total**: ~100KB of documentation

No caching needed - regeneration is cheap.

## Advanced Usage

### Programmatic Usage

Use the generator as a library:

```go
import (
    "github.com/powertoolsdev/mono/pkg/gen/markdown-docs/mdast"
    "github.com/powertoolsdev/mono/pkg/config/schema"
)

// Create custom documentation
doc := mdast.NewDocument()
doc.AddHeading(1, "Custom Doc")

// Use schema data
s, _ := schema.LookupSchemaType("action")
// ... process schema ...

markdown := doc.Render()
```

### Batch Processing

Generate for specific schemas only:

```go
// Modify main.go to accept schema filter
schemasToGenerate := []string{"action", "helm", "terraform"}
for _, name := range schemasToGenerate {
    generateSchemaDoc(name, outputDir, format)
}
```

## Related Documentation

- [README.md](README.md) - Architecture and design
- [mdast/ast.go](mdast/ast.go) - AST node types
- [pkg/config/schema/types.go](../../config/schema/types.go) - Schema definitions
- [API endpoint](../../../services/ctl-api/internal/app/general/service/config_schema.go) - REST API
