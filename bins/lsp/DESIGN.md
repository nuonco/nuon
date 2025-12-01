# Nuon LSP Design Document

This document describes the key architectural decisions made in the Nuon Language Server Protocol implementation.

## 1. Custom TOML Parser Instead of Tree-Sitter

### Problem

Tree-sitter is a popular choice for language parsing in LSP implementations, but it has a critical limitation for this use case: **CGo bindings**. Including tree-sitter would introduce:
- C/C++ dependencies that complicate the build process
- Additional compilation steps and system dependencies
- Complexity in distribution and deployment
- Potential portability issues across platforms

For an LSP that needs to be lightweight and easy to develop, these trade-offs weren't acceptable.

### Solution

The LSP uses a **custom loose TOML parser** implemented in `pkg/parser/toml/` with two parsing modes:

#### Loose Parser (for LSP Features)

The loose parser in `loose.go` uses regex patterns and never returns errors:

```go
var (
    tableHeaderRegex      = regexp.MustCompile(`^\s*\[\s*([A-Za-z0-9_.-]+)`)
    keyValueRegex         = regexp.MustCompile(`^\s*([A-Za-z0-9_.-]+)\s*=?`)
    arrayTableHeaderRegex = regexp.MustCompile(`^\s*\[\[\s*([A-Za-z0-9_.-]+)`)
)
```

**Key characteristics:**
- **Never errors on incomplete input** - Critical for LSP where users are actively typing
- **Preserves position information** - Each parsed element includes `Range` with `Start` and `End` positions
- **Regex-based extraction** - Fast and simple, sufficient for extracting table headers, key names, and basic structure
- **Cursor-aware parsing** - `ParseLooseWithCursor()` detects the prefix at a specific cursor position

#### Strict Parser (for Validation)

Used only for validation/diagnostics:

```go
func ValidateToml(text string) error {
    _, err := ParseStrict(text)
    return err
}
```

This allows the LSP to validate syntax separately from providing features.

#### Usage in LSP

The completion handler uses the loose parser:

```go
cursorPos := tomlparser.Position{Line: int(pos.Line), Character: int(pos.Character)}
doc := tomlparser.ParseTomlWithCursor(text, cursorPos)
tomlCtx := doc.ContextAt(cursorPos)
```

The loose parser provides:
- Accurate line/column tracking for completion positions
- Graceful handling of incomplete TOML while the user is typing
- Extracted table context and key information for context-aware completions

### Trade-offs

| Aspect | Loose Parser | Tree-Sitter |
|--------|--------------|-------------|
| Dependencies | None (stdlib only) | CGo bindings |
| Incomplete input | ✅ Graceful | ❌ Errors |
| Position tracking | ✅ Explicit tracking | ✅ Built-in |
| Build complexity | ✅ Simple | ❌ Complex |
| Coverage | Basic TOML | Full TOML spec |

For the Nuon LSP use case (TOML application manifests), the loose parser provides sufficient parsing capability while maintaining simplicity.

## 2. Hierarchical JSON Schema Property Mapping

### Problem

The LSP needs to provide **context-aware completions**. When a user is editing inside a specific TOML table section, the LSP should only suggest properties that are valid at that nesting level.

For example, in:
```toml
[public_repo]
username = "..."

[private_repo]
path = "..."
```

When cursor is in `[public_repo]`, suggest only properties valid for `public_repo`.
When cursor is in `[private_repo]`, suggest only properties valid for `private_repo`.

A naive approach (flattening all schema properties) would suggest all properties everywhere, breaking the context-awareness.

### Solution

The LSP builds a **hierarchical property map** in `mappers/schema.go`:

```go
func BuildPropertyMap(schema *jsonschema.Schema) map[string]map[string]*jsonschema.Schema {
    hierarchicalMap := make(map[string]map[string]*jsonschema.Schema)
    // ...
}
```

**Structure:**
- **Outer map key**: Table path in dotted notation (e.g., `""` for root, `"public_repo"`, `"public_repo.auth"`)
- **Inner map**: Properties available at that level
- **Inner map values**: Full schema information (type, enum, examples, description)

#### How it Works

The recursive function `buildPropertyMapRecursive()` traverses the JSON schema:

1. **Direct properties** are stored at the current level
2. **Nested objects** via `$ref` or inline `properties` create new entries with dotted paths
3. **Array items** are handled similarly - if an array contains an object, that object's properties become a new level

Example mapping for a schema with tables:
```
{
    "": {
        "type": SchemaNode,
        ...
    },
    "public_repo": {
        "username": SchemaNode,
        "token": SchemaNode,
        ...
    },
    "private_repo": {
        "path": SchemaNode,
        ...
    }
}
```

#### $ref Resolution

The mapper resolves `$ref` pointers to schema definitions:

```go
if prop.Ref != "" {
    refDef := resolveRef(prop.Ref, defsLookup)
    if refDef != nil {
        nestedPath := key
        if currentPath != "" {
            nestedPath = currentPath + "." + key
        }
        buildPropertyMapRecursive(refDef, nestedPath, hierarchicalMap, defsLookup)
    }
}
```

This allows the schema to define reusable types via `$ref` without losing the hierarchical structure.

#### Usage in Completions

In the completion handler:

```go
propertiesAtLevel, ok := hierarchicalMap[tomlCtx.CurrentTable]
if !ok || len(propertiesAtLevel) == 0 {
    log.Warningf("No properties found for table '%s'", tomlCtx.CurrentTable)
} else {
    for key, prop := range propertiesAtLevel {
        // Generate completion for this property at this level
    }
}
```

The completion handler looks up the current table in the hierarchical map and only suggests properties available at that level.

### Benefits

- **Context awareness**: Completions are filtered to the current nesting level
- **Schema-driven**: No hard-coded property lists; all data comes from the schema
- **Nested structure support**: Properly handles arbitrary nesting depths
- **Type information**: Each property includes its type, enum values, and examples
- **Reusability**: $ref definitions are shared across multiple locations

## 3. Context Detection for Completions

### Problem

The LSP needs to determine:
1. **Where is the cursor?** (on a key or on a value)
2. **What table are we in?** (current TOML table context)
3. **What key is being completed?** (if on a value)

This information drives the completion strategy.

### Solution

The completion handler detects context in multiple stages:

#### Stage 1: Parse with Cursor

```go
cursorPos := tomlparser.Position{Line: int(pos.Line), Character: int(pos.Character)}
doc := tomlparser.ParseTomlWithCursor(text, cursorPos)
tomlCtx := doc.ContextAt(cursorPos)
```

The loose parser extracts cursor-aware context.

#### Stage 2: Analyze Current Line

```go
lines := strings.Split(text, "\n")
currentLine := ""
if int(pos.Line) < len(lines) {
    currentLine = lines[pos.Line]
}
beforeCursor := currentLine[:min(int(pos.Character), len(currentLine))]
isValue := strings.Contains(beforeCursor, "=")
```

Check if `=` appears before the cursor - if yes, we're completing a value; otherwise, a key.

#### Stage 3: Route to Appropriate Completion Logic

**For key completions** (no `=` before cursor):

```go
if !isValue {
    propertiesAtLevel, ok := hierarchicalMap[tomlCtx.CurrentTable]
    if !ok || len(propertiesAtLevel) == 0 {
        // No properties at this level
    } else {
        for key, prop := range propertiesAtLevel {
            items = append(items, protocol.CompletionItem{
                Label:         key,
                Detail:        &prop.Type,
                Documentation: &protocol.MarkupContent{...},
                InsertText:    ptr(fmt.Sprintf("%s = ", key)),
            })
        }
    }
}
```

Suggests property names available at the current table level, with auto-inserted `= ` for convenience.

**For value completions** (`=` before cursor):

```go
else if tomlCtx.KeyOnLine != "" {
    propertiesAtLevel, ok := hierarchicalMap[tomlCtx.CurrentTable]
    prop, ok := propertiesAtLevel[tomlCtx.KeyOnLine]
    
    if prop.Type == "boolean" {
        items = append(items,
            protocol.CompletionItem{Label: "true"},
            protocol.CompletionItem{Label: "false"},
        )
    } else if len(prop.Enum) > 0 {
        // Suggest enum values with proper quoting for strings
        for _, enumVal := range prop.Enum {
            value := fmt.Sprintf("%v", enumVal)
            if isStringType {
                value = fmt.Sprintf("\"%s\"", value)
            }
            items = append(items, protocol.CompletionItem{...})
        }
    } else if len(prop.Examples) > 0 {
        // Suggest example values
    }
}
```

Suggests appropriate values based on the property's type:
- Boolean → `true` / `false`
- Enum → enum values
- Other types → example values
- String values are automatically quoted

### Key Insights

1. **Position-aware parsing**: The loose parser preserves positions, enabling precise cursor-relative logic
2. **Dual-mode logic**: Same infrastructure supports both key and value completions
3. **Schema-driven suggestions**: All suggestions come from schema type/enum/examples
4. **Proper quoting**: String values are automatically wrapped in quotes

## Architecture Summary

```
┌─────────────────────────────────────┐
│   LSP Request Handler               │
│   (handlers/language-features.go)   │
└────────────┬────────────────────────┘
             │
    ┌────────┼─────────┐
    │        │         │
    ▼        ▼         ▼
┌────────┐ ┌────────────────┐ ┌──────────────┐
│ Custom │ │ Hierarchical   │ │ Context      │
│ Loose  │ │ Schema         │ │ Detection    │
│ Parser │ │ Property Map   │ │              │
└────────┘ └────────────────┘ └──────────────┘
    │              │                 │
    │              ▼                 │
    │      ┌──────────────┐         │
    │      │ JSON Schema  │         │
    │      │ Definitions  │         │
    │      └──────────────┘         │
    │                               │
    └───────────────┬───────────────┘
                    │
                    ▼
        ┌──────────────────────┐
        │ Completion Items     │
        │ Hover Information    │
        └──────────────────────┘
```

The three design choices work together:

1. **Custom parser** provides position-aware TOML structure
2. **Hierarchical map** enables schema-driven, context-aware suggestions
3. **Context detection** routes to the right completion logic

This architecture keeps the LSP simple, dependency-free, and focused on the specific needs of editing Nuon configuration files.
