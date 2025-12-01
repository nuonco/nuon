# Nuon Language Server Protocol (LSP) - Developer Guide

This document provides comprehensive guidance for AI coding agents working on the Nuon LSP project.

## Project Overview

The Nuon LSP is a Language Server Protocol implementation for editing Nuon configuration files (TOML format). It provides:

- **Code Completion** - Context-aware suggestions with trigger characters (`=` and space)
- **Hover Information** - Contextual documentation on hover
- **Document Synchronization** - Full document text tracking
- **No external binary dependencies** - Pure Go with minimal stdlib usage (no tree-sitter, no CGo)

## Architecture

The LSP is built on three key architectural decisions:

### 1. Custom Loose TOML Parser (`pkg/parser/toml/`)

Instead of tree-sitter (which requires CGo bindings), the LSP implements a custom TOML parser that:
- **Never errors** on incomplete input (essential for real-time editing)
- **Preserves position information** (line/column for each element)
- **Uses regex patterns** for fast, lightweight parsing
- **Has dual modes**: loose (for LSP features) and strict (for validation)

**Files:**
- `pkg/parser/toml/parser.go` - Public API and documentation
- `pkg/parser/toml/loose.go` - Loose parser implementation with regex patterns
- `pkg/parser/toml/strict.go` - Strict parser for validation (using a proper TOML library)

**Key insight:** The parser never fails on incomplete TOML, which is crucial because the LSP operates on documents as users are actively typing.

### 2. Hierarchical JSON Schema Property Mapping (`bins/lsp/mappers/schema.go`)

The LSP maps JSON schemas into a hierarchical structure that enables context-aware completions:

```go
map[string]map[string]*jsonschema.Schema
// Key: table path (e.g., "" for root, "public_repo", "public_repo.auth")
// Value: properties available at that level
```

**Benefits:**
- When cursor is in `[public_repo]` table, only suggest properties for that table
- Resolves `$ref` pointers to schema definitions for reusable types
- Supports arbitrary nesting depth via dotted path notation
- Full schema information (type, enum, examples) available for each property

**Files:**
- `bins/lsp/mappers/schema.go` - BuildPropertyMap function and recursive traversal

### 3. Context Detection for Completions (`bins/lsp/handlers/completion.go`)

The completion handler detects:
- **Cursor position** - Are we on a key or value?
- **Current table context** - What TOML table are we in?
- **Active key** - Which property is being edited?

This detection routes to the appropriate completion logic:
- **Key completions**: Suggest property names from schema at current table level
- **Value completions**: Suggest enum values, booleans, or examples for the active key

## Directory Structure

```
bins/lsp/
├── main.go                      # LSP initialization and handler registration
├── handlers/
│   ├── completion.go            # TextDocumentCompletion handler
│   ├── did_change.go            # TextDocumentDidChange handler + helpers
│   ├── did_open_close.go        # TextDocumentDidOpen/DidClose handlers
│   ├── hover.go                 # TextDocumentHover handler + HoverProvider
│   ├── helpers.go               # Shared types, utilities, and logger
│   ├── opendocs.go              # Open documents tracking (openDocuments map)
│   └── language_features_test.go # Handler tests
├── mappers/
│   └── schema.go                # Hierarchical schema property mapping
├── models/
│   └── schema_lookup.go         # Schema type detection and lookup
├── nuon-lsp-vscode/             # VSCode extension
│   ├── extension.js             # Extension activation and LSP client setup
│   └── package.json             # Extension metadata and dependencies
├── README.md                    # User-facing development guide
├── DESIGN.md                    # Architectural decision documentation
└── AGENTS.md                    # This file
```

## Key Files Explained

### `main.go`

Entry point for the LSP:
- Initializes commonlog for logging
- Registers handler functions with the glsp library
- Configures server capabilities (completion, hover, text sync)
- Starts stdio-based server

```go
handler = protocol.Handler{
    Initialize:             initialize,
    Shutdown:               shutdown,
    TextDocumentCompletion: handlers.TextDocumentCompletion,
    TextDocumentDidOpen:    handlers.TextDocumentDidOpen,
    TextDocumentDidChange:  handlers.TextDocumentDidChange,
    TextDocumentDidClose:   handlers.TextDocumentDidClose,
    TextDocumentHover:      handlers.TextDocumentHover,
}
```

### `handlers/completion.go`

**`TextDocumentCompletion`**: Main completion logic
1. Parse document with cursor position using loose parser
2. Detect context (current table, key/value)
3. Look up schema for the detected schema type
4. Build hierarchical property map
5. Route to key or value completion logic
6. Return completion items with proper quoting/formatting

### `handlers/hover.go`

**`TextDocumentHover`**: Provides documentation on hover
- Looks up property schema at cursor position
- Returns markdown-formatted documentation

**`HoverProvider`**: Reusable helper for hover requests
- Navigates hierarchical schema to find property at cursor
- Formats documentation with description, enum values, and examples

### `handlers/did_open_close.go`

**`TextDocumentDidOpen`**: Opens a document
- Tracks document in `openDocuments` map for quick lookup

**`TextDocumentDidClose`**: Closes a document
- Removes document from `openDocuments` map

### `handlers/did_change.go`

**`TextDocumentDidChange`**: Updates document text on user edits
- Handles both full document sync and incremental changes
- Supports range-based text updates with `applyTextChange()`

**`extractChangeContent()`**: Utility to parse change events
- Handles both typed structs and JSON unmarshaled maps
- Extracts text and range information

**`applyTextChange()`**: Applies incremental text changes
- Single-line and multi-line change support
- Maintains line/column integrity

### `handlers/helpers.go`

**`Document`**, **`Range`**, **`Context`**: Type definitions
- Shared structures used across handlers

**`detectContext()`**: Context detection helper
- Finds current table and key at cursor position
- Determines if cursor is on key or value

**Logger**: Initialized with `lsp-handlers` tag for structured logging

### `mappers/schema.go`

Core algorithm for hierarchical property mapping:

```go
func BuildPropertyMap(schema *jsonschema.Schema) map[string]map[string]*jsonschema.Schema
```

**Process:**
1. Start with root schema, resolve any `$ref` at root level
2. For each property in current level:
   - Store property at current path level
   - If property has `$ref`: recursively process referenced definition
   - If property is array with items: recursively process array item schema
   - If property has inline `properties`: recursively process inline object
3. Build paths using dotted notation (e.g., "public_repo.auth.username")

**Important:** The function uses `resolveRef()` to handle JSON Schema `$ref` pointers, supporting both `#/definitions/TypeName` and `#/$defs/TypeName` formats.

### `models/schema_lookup.go`

Schema detection and lookup:

**`DetectSchemaType(text string) string`**: Finds the `type = "..."` key in TOML
- Scans first few lines of document
- Stops at first table section `[table]`
- Returns the detected schema type (e.g., "helm")

**`LookupSchema(schemaType string) (*jsonschema.Schema, error)`**: Retrieves schema
- Delegates to `pkg/config/schema.LookupSchemaType()`
- Returns parsed JSON schema for the type

### `nuon-lsp-vscode/extension.js`

VSCode extension that connects to the LSP server:

```javascript
const config = vscode.workspace.getConfiguration("nuonLsp");
let serverCommand = config.get("serverPath");
if (!serverCommand || serverCommand === "") {
    serverCommand = path.join(context.extensionPath, "..", "lsp");
}

const serverOptions = {
    command: serverCommand,
    transport: TransportKind.stdio,
};
const clientOptions = {
    documentSelector: [{ scheme: "file", language: "toml" }],
};
```

**Configuration:** Set `nuonLsp.serverPath` in VSCode settings to point to your LSP binary, or leave empty to use the bundled binary.

**Important:** Must be opened as workspace root in VSCode for debugging to work (see README.md).

## Development Workflow

### Building the LSP

```bash
cd bins/lsp
go build -o lsp ./
```

Creates an executable binary that communicates via stdio.

### Testing in Editors

#### Neovim

1. Build the binary
2. Update Neovim config with path to binary
3. Reload config: `:source $MYVIMRC`
4. Check status: `:LspInfo`
5. View logs: `:LspLog`

#### VSCode

1. Build the binary
2. Update `extension.js` with path to binary
3. Open `nuon-lsp-vscode/` folder as workspace root
4. Press `F5` to launch Extension Development Host
5. Check "Nuon LSP" output panel for logs

### Debugging

**Enable debug logging in main.go:**
```go
commonlog.Configure(2, nil)  // Change verbosity level (0=high, 2=low)
```

**View logs:**
- VSCode: Output panel → "Nuon LSP" dropdown
- Neovim: `:LspLog` command

### Testing Completions Locally

1. Build the LSP binary
2. Create a test TOML file with `type = "helm"` at the top
3. Open in editor with LSP configured
4. Position cursor in `[table_name]` section
5. Type `Ctrl+Space` (or editor's completion trigger)
6. Verify suggestions appear for that table's properties

## Code Patterns

### Adding a New Handler

1. Create a new handler file in `handlers/` directory (e.g., `handlers/new_feature.go`):
   ```go
   package handlers

   import (
       "github.com/tliron/glsp"
       protocol "github.com/tliron/glsp/protocol_3_16"
   )

   func TextDocumentNewFeature(ctx *glsp.Context, params *protocol.NewFeatureParams) (any, error) {
       log.Infof("Feature triggered")
       // Implementation
       return result, nil
   }
   ```

2. Register in `main.go`:
   ```go
   handler = protocol.Handler{
       // ... existing handlers ...
       TextDocumentNewFeature: handlers.TextDocumentNewFeature,
   }
   ```

3. Update `initialize()` in `main.go` to advertise capability:
   ```go
   return protocol.InitializeResult{
       Capabilities: protocol.ServerCapabilities{
           // ... existing capabilities ...
           NewFeature: true,
       },
   }, nil
   ```

**File Organization Tips:**
- Each handler should have its own file (e.g., `completion.go`, `hover.go`)
- Helper functions specific to a handler should live in the same file
- Shared utilities go in `helpers.go`
- Document tracking state lives in `opendocs.go`

### Adding a New Schema Type

1. Create schema definition in the appropriate location (see `pkg/config/schema`)
2. Define JSON schema for the type
3. Update `DetectSchemaType()` if detection logic changes
4. LSP will automatically:
   - Look up the schema via `LookupSchema()`
   - Build hierarchical property map
   - Provide completions based on schema

### Improving Completion Accuracy

The completion logic has multiple debug points:

```go
log.Infof("📝 Completion requested at %s:%d:%d", uri, pos.Line, pos.Character)
log.Debugf("✅ Found document, length: %d chars", len(text))
log.Debugf("✅ TOML parsed, %d tables, %d keys found", len(doc.Tables), len(doc.Keys))
log.Infof("📍 Context detected - Table: '%s', CurrentKey: '%s'", 
    tomlCtx.CurrentTable, tomlCtx.KeyOnLine)
log.Debugf("✅ Detected schema type: %s", schemaType)
log.Infof("✅ Generated %d key completions", count)
```

Enable debug logging to trace completion execution.

## Dependencies

### Go Libraries

- **`github.com/tliron/glsp`** - LSP protocol implementation
- **`github.com/tliron/commonlog`** - Structured logging
- **`github.com/invopop/jsonschema`** - JSON schema parsing and traversal
- **`github.com/powertoolsdev/mono/pkg/config/schema`** - Schema lookup (internal)
- **`github.com/powertoolsdev/mono/pkg/parser/toml`** - Custom TOML parser (internal)

### Node.js (for VSCode extension)

- **`vscode`** - VSCode extension API
- **`vscode-languageclient`** - LSP client for VSCode
- Built with webpack for distribution

## Important Notes

### Position Tracking

The loose TOML parser explicitly tracks positions because:
1. LSP requires precise line/column information for features like completion
2. Strict parsers often lose position data
3. Real-time editing means documents are incomplete during typing

Always use `ParseTomlWithCursor()` for features that depend on cursor position.

### Schema Type Detection

The LSP detects schema type from the first `type = "..."` key in the TOML:

```toml
# Must appear before any [table] headers
type = "helm"

[public_repo]
username = "..."
```

If no type is detected, the LSP returns no completions (logs warning).

### Error Handling in LSP

The loose parser **never returns errors**, so LSP handlers must:
1. Check for nil values (missing schema, empty property maps)
2. Log warnings for missing context
3. Gracefully degrade (return empty completions instead of error)

This is intentional - the LSP must remain operational even with malformed input.

### String Quoting in Completions

Value completions automatically handle quoting:

```go
if isStringType {
    value = fmt.Sprintf("\"%s\"", value)
}
```

This ensures users get valid TOML syntax without manual quoting.

## Testing Strategy

### Unit Tests

Located in `handlers/language_features_test.go`. To add new tests:

1. Tests are organized by handler functionality
2. Mock `openDocuments` or refactor to inject dependency
3. Test completion logic with various table contexts
4. Test context detection with edge cases

### Integration Tests

Test the full LSP with a real editor:

1. Use Neovim or VSCode extension
2. Create test TOML files with different schemas
3. Verify completions appear at expected locations
4. Check hover information formatting

### Manual Testing Checklist

- [ ] Completion triggers on `=` character
- [ ] Completion triggers on space character
- [ ] Completion items have correct `InsertText` (with `= ` for keys)
- [ ] String values are quoted in completions
- [ ] Boolean values show `true`/`false`
- [ ] Enum values are properly formatted
- [ ] Hover shows documentation for properties
- [ ] No errors when TOML is incomplete
- [ ] Document sync works after multiple edits

## Common Issues and Solutions

### LSP Won't Start

**Check:**
1. Binary path is correct in editor config
2. Binary is executable: `chmod +x bins/lsp/lsp`
3. Build succeeded: `go build -o lsp ./`

**Debug:**
```bash
# Try running manually to see stderr
./lsp
# Should wait for input (it's a stdio server)
# Ctrl+C to exit
```

### No Completions Showing

**Check:**
1. File is TOML format (editor.documentSelector must match)
2. TOML has `type = "..."` declaration
3. LSP can find the schema for that type
4. Cursor is in a recognized table section

**Debug:**
Enable debug logging and check output for:
- "Schema type detected: ..."
- "Context detected - Table: ..."
- "Generated X key/value completions"

### VSCode Extension Won't Activate

**Check:**
1. Opened `nuon-lsp-vscode/` folder as workspace root (not parent folder)
2. Ran `npm install` in that folder
3. Binary path in `extension.js` is correct
4. Debug Console shows activation message

**Debug:**
Press `F5` with `extension.js` open to set breakpoints and step through activation.

## References

- **LSP Specification**: https://microsoft.github.io/language-server-protocol/
- **glsp Library**: https://github.com/tliron/glsp
- **JSON Schema**: https://json-schema.org/
- **VSCode Extension API**: https://code.visualstudio.com/docs/extensions/language-support
- **Nuon LSP README**: README.md (user-facing guide)
- **Nuon LSP DESIGN**: DESIGN.md (architectural decisions)

## Contributing Guidelines

When making changes to the LSP:

1. **Follow the custom TOML parser design**: Don't add external parsing dependencies
2. **Maintain position tracking**: Position information is critical for LSP features
3. **Update hierarchical map logic carefully**: This is the core of context awareness
4. **Test with both editors**: Changes should work in both Neovim and VSCode
5. **Add logging**: Use the commonlog patterns already established
6. **Document changes**: Update README.md, DESIGN.md, or this file as needed

### Code Style

- Follow Go conventions (use `gofmt`)
- Use descriptive function names
- Include structured logging with emoji indicators (already established pattern)
- Write comments for non-obvious logic

### Performance Considerations

The LSP should be responsive for real-time editing:
- Completion should return results <100ms
- Hover should return results <100ms
- Document parsing should handle large files efficiently

The loose regex-based parser is designed to be fast for typical TOML sizes (under 10KB).
