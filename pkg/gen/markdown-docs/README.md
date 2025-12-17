# Markdown Documentation Generator

A `go generate` compatible tool for generating markdown documentation from Nuon configuration JSON schemas.

## Overview

This package provides a sophisticated markdown AST (Abstract Syntax Tree) generator that produces high-quality documentation from JSON schemas. It mirrors how the Nuon API exposes schemas via the `/v1/general/config-schema` endpoint.

## Architecture

The generator follows the same pattern as the API:

1. **Schema Resolution**: Uses `schema.LookupSchemaType()` to retrieve schemas, identical to the API endpoint in `services/ctl-api/internal/app/general/service/config_schema.go`

2. **Schema Processing**: Handles `$ref` definitions and nested schemas just like the API returns them

3. **Markdown AST**: Generates a structured AST that can be rendered to different markdown formats

## Components

### Main Generator (`main.go`)
- Command-line tool for generating documentation
- Supports multiple output formats (Mintlify, GitHub, plain)
- Configurable output directory
- Verbose logging option

### Markdown AST (`mdast/ast.go`)
- Type-safe markdown document structure
- Composable node types (Heading, Table, CodeBlock, etc.)
- MDX/JSX character escaping for Mintlify compatibility
- Renders to clean, standards-compliant markdown

## Usage

### Via go generate

Add a `gen.go` file to your schema package:

```go
package schema

//go:generate go run github.com/nuonco/nuon/pkg/gen/markdown-docs -output=../../../docs/config-ref -format=mintlify
```

Then run:

```bash
go generate ./pkg/config/schema
```

### Direct Execution

```bash
# Generate Mintlify docs (default)
go run ./pkg/gen/markdown-docs -output=docs/config-ref -format=mintlify

# Generate GitHub-flavored markdown
go run ./pkg/gen/markdown-docs -output=docs/config-ref -format=github

# Enable verbose logging
go run ./pkg/gen/markdown-docs -output=docs/config-ref -verbose
```

## Command-Line Flags

- `-output` - Output directory for generated files (default: `docs/config-ref`)
- `-format` - Output format: `mintlify`, `github`, or `plain` (default: `mintlify`)
- `-verbose` - Enable verbose logging

## Output Structure

The generator creates:

1. **Individual Schema Files**: One `.mdx` file per schema type
   - Frontmatter with title and description (Mintlify format)
   - Properties table with type, required, description, default, example
   - Property details section for enums and multiple examples

2. **Index Page**: Navigation page with categorized schemas
   - Component Types (helm, terraform, docker-build, etc.)
   - Configuration Types (inputs, secrets, policies, etc.)
   - Other Types (action, install, runner, etc.)

## Schema Resolution

The generator uses the same schema resolution logic as the API:

```go
// 1. Lookup schema by type (same as API endpoint)
s, err := schema.LookupSchemaType(schemaName)

// 2. Resolve $ref definitions (handles nested schemas)
targetSchema := resolveSchemaRef(s)

// 3. Generate markdown from properties
// (iterates through Properties.Oldest() like the API would)
```

## Markdown AST Features

### Document Structure
```go
doc := mdast.NewDocument()
doc.AddFrontmatter(map[string]string{
    "title": "Example",
    "description": "Example description",
})
doc.AddHeading(1, "Title")
doc.AddParagraph("Some text")
doc.AddTable(table)
```

### Table Generation
```go
table := mdast.NewTable([]string{"Property", "Type", "Required"})
table.AddRow([]string{
    mdast.Code("name"),
    mdast.Code("string"),
    "✅ Yes",
})
doc.AddTable(table)
```

### MDX Safety
```go
// Automatically escapes special characters
desc := mdast.EscapeMDX("Use {{.nuon.install.id}} for templating")
// Output: "Use \\{\\{.nuon.install.id\\}\\} for templating"
```

## Integration with API

This generator mirrors the API's `/v1/general/config-schema` endpoint:

**API Endpoint** (`services/ctl-api/internal/app/general/service/config_schema.go`):
```go
func (s *service) GetConfigSchema(ctx *gin.Context) {
    typ := ctx.DefaultQuery("type", "")
    schm, err := schema.LookupSchemaType(typ)  // Same lookup!
    ctx.JSON(http.StatusOK, schm)
}
```

**Generator**:
```go
func generateSchemaDoc(schemaName string) error {
    s, err := schema.LookupSchemaType(schemaName)  // Same lookup!
    // ... generate markdown from s ...
}
```

This ensures documentation always matches what the API returns.

## Extending the Generator

### Adding New Output Formats

1. Add format to `categorizeSchemas()` logic
2. Update `generateIndexPage()` to handle new format
3. Extend markdown AST if needed for format-specific features

### Custom AST Nodes

Add new node types to `mdast/ast.go`:

```go
type CustomNode struct {
    // fields
}

func (c *CustomNode) Render() string {
    // rendering logic
}
```

### Schema Processing Hooks

Extend `generateSchemaDoc()` to add custom processing:

```go
// Add custom sections
if schemaName == "special-type" {
    doc.AddHeading(2, "Special Section")
    // custom logic
}
```

## Future Enhancements

- [ ] Support for schema inheritance visualization
- [ ] Interactive example generation
- [ ] Schema validation examples
- [ ] Link generation between related schemas
- [ ] OpenAPI/Swagger output format
- [ ] JSON Schema validation integration

## Maintenance

When schemas change:

1. Run `go generate ./pkg/config/schema`
2. Review generated markdown for quality
3. Commit both schema changes and generated docs

## Related Files

- `pkg/config/schema/types.go` - Schema definitions and mapping
- `services/ctl-api/internal/app/general/service/config_schema.go` - API endpoint
- `docs/config-ref/` - Generated documentation output
