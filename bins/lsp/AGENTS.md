# Nuon LSP — Agent Guide

Language Server Protocol server for Nuon TOML configuration files. Pure Go (no CGo, no tree-sitter). Communicates over stdio in production; optional TCP and health-check HTTP ports for local development.

## Capabilities

| Feature | Handler | Notes |
|---------|---------|-------|
| Completion | `handlers/completion.go` | Trigger chars: `=`, space, `#` |
| Hover | `handlers/hover.go` | Markdown docs from JSON Schema |
| Diagnostics | `handlers/diagnostics.go` | Schema validation, unknown keys, required fields |
| Formatting | `handlers/formatting.go` | Opinionated TOML layout (alignment, blank lines) |
| Folding ranges | `handlers/folding_range.go` | Tables, multiline strings, comment blocks |
| Document sync | `handlers/did_*.go` | Full sync; in-memory `openDocuments` map |
| Workspace scan | `handlers/workspace.go` | On init, publishes diagnostics for all `.toml` files |

## Directory Structure

```
bins/lsp/
├── main.go                         # Server init, handler registration, transport modes
├── Dockerfile                      # Multi-stage build (test, lint, cross-platform artifacts)
├── handlers/
│   ├── handlers.go                 # Package logger (`lsp-handlers`)
│   ├── opendocs.go                 # `openDocuments` map + RWMutex
│   ├── did_open.go                 # Open doc, publish diagnostics
│   ├── did_change.go               # Full/incremental sync, publish diagnostics
│   ├── did_close.go                # Remove from openDocuments
│   ├── did_save.go                 # Refresh cache, publish diagnostics
│   ├── completion.go               # TextDocumentCompletion
│   ├── hover.go                    # TextDocumentHover + HoverProvider
│   ├── diagnostics.go              # PublishDiagnostics, DiagnoseDocument
│   ├── formatting.go               # TextDocumentFormatting, FormatToml
│   ├── folding_range.go            # TextDocumentFoldingRange
│   ├── workspace.go                # Workspace folder tracking, TOML scan
│   ├── diagnostics_test.go
│   ├── formatting_test.go
│   └── folding_range_test.go
├── mappers/
│   ├── schema.go                   # BuildPropertyMap (hierarchical schema → table paths)
│   └── schema_test.go
├── models/
│   ├── schema_lookup.go            # DetectSchemaType, LookupSchema, validation helpers
│   └── schema_lookup_test.go
├── nuon-lsp-vscode/                # VS Code extension (plain JS, no bundler)
│   ├── extension.js
│   └── package.json
├── README.md                       # User-facing install and usage
├── DESIGN.md                       # Architecture rationale
└── AGENTS.md                       # This file
```

Shared libraries outside this directory:

- `pkg/parser/toml/` — loose parser (LSP features), strict parser (validation)
- `pkg/config/schema/` — embedded JSON schemas and lookup

## Architecture

### 1. Loose TOML parser (`pkg/parser/toml/`)

The LSP cannot rely on strict parsers that fail on incomplete input.

- `ParseToml` / `ParseTomlWithCursor` — always use loose parsing; never error; preserve line/column positions
- `ValidateToml` — strict parse for diagnostics only
- `doc.ContextAt(cursorPos)` — table path, key on line, key path (replaces any legacy `detectContext` helper)

Use `ParseTomlWithCursor` for completion and hover. Use loose parse + `ValidateToml` / schema checks for diagnostics.

### 2. Schema type detection (`models/schema_lookup.go`)

Schema type comes from the **first non-empty comment line** at the top of the file:

```toml
#helm

[public_repo]
username = "..."
```

`DetectSchemaType` scans leading blank lines and comment lines; it stops at the first non-comment line. If no comment is found, completions and most diagnostics are skipped.

Completion also offers schema type names when the cursor is in the top comment block and typing after `#` (trigger char `#`).

### 3. Hierarchical schema mapping (`mappers/schema.go`)

```go
func BuildPropertyMap(schema *jsonschema.Schema) (
    map[string]map[string]*jsonschema.Schema,  // table path → properties
    map[string]map[string]bool,                // table path → required field names
)
```

- Outer key: TOML table path (`""` for root, `"public_repo"`, `"public_repo.auth"`, …)
- Resolves `$ref` via `#/definitions/` and `#/$defs/`
- Recurses into inline objects, array items, and `allOf` branches

Completion and hover look up properties at `tomlCtx.CurrentTable`. Hover also surfaces required-field status from `requiredMap`.

### 4. Document state and diagnostics flow

1. `did_open` / `did_change` / `did_save` update `openDocuments[uri]`
2. Each update calls `PublishDiagnostics`
3. On initialize, `ScanWorkspaceForDiagnostics` walks workspace folders and publishes diagnostics for every `.toml` file (even if not open)

Diagnostics cover: unknown schema type, schema lookup failures, missing required fields, unknown keys, type mismatches. Invalid TOML still parses loosely; strict validation supplements where needed.

## `main.go` essentials

```go
handler = protocol.Handler{
    TextDocumentCompletion:   handlers.TextDocumentCompletion,
    TextDocumentDidOpen:      handlers.TextDocumentDidOpen,
    TextDocumentDidChange:    handlers.TextDocumentDidChange,
    TextDocumentDidClose:     handlers.TextDocumentDidClose,
    TextDocumentDidSave:      handlers.TextDocumentDidSave,
    TextDocumentHover:        handlers.TextDocumentHover,
    TextDocumentFoldingRange: handlers.TextDocumentFoldingRange,
    TextDocumentFormatting:   handlers.TextDocumentFormatting,
}
```

Capabilities advertised: full document sync, hover, completion (triggers `=`, ` `, `#`), folding ranges, document formatting.

Transport flags:

- Default (no flags): stdio — used by VS Code extension and most editors
- `-port N`: TCP on `127.0.0.1:N` for dev attach
- `-health-port N`: HTTP `/health` endpoint

Logging: `commonlog.Configure(2, nil)` — lower number = more verbose.

## Development

### Build and test

```bash
cd bins/lsp
go build -o nuon-lsp ./
go test ./...
```

After Go edits, run `gofmt` and `goimports` on changed packages.

### TCP dev loop (VS Code)

Terminal 1:

```bash
go build -o nuon-lsp ./ && ./nuon-lsp -port 8765
```

VS Code setting `"nuonLsp.port": 8765` makes the extension connect via TCP instead of spawning stdio. Useful for breakpoints in the server.

### Stdio mode (VS Code)

Extension resolves server binary in order:

1. `nuonLsp.serverPath` setting
2. `nuon-lsp` on `PATH`
3. `../lsp` relative to extension directory (local dev fallback)

Open `nuon-lsp-vscode/` as the VS Code workspace root and press F5 for Extension Development Host.

### Neovim

Build the binary, point `cmd` at it in LSP config, reload, verify with `:LspInfo`, logs via `:LspLog`.

## Adding a handler

1. Add `handlers/new_feature.go` with a `TextDocumentNewFeature` function
2. Register on `protocol.Handler` in `main.go`
3. Advertise the capability in `initialize()`
4. Keep handler-specific helpers in the same file; shared state stays in `opendocs.go` or `workspace.go`

One handler per file is the established pattern.

## Adding a schema type

1. Add the JSON schema under `pkg/config/schema`
2. Register the type name in schema lookup helpers there
3. LSP picks it up automatically via `models.LookupSchema` — no handler changes unless detection rules change

## Conventions worth preserving

**Graceful degradation.** Loose parsing never errors. Handlers return empty results (not errors) when schema type is missing or context is ambiguous.

**Concurrency.** Always lock `openDocumentsMutex` when reading or writing `openDocuments`.

**String quoting.** Value completions quote strings so inserted text is valid TOML.

**Key completions.** Insert text includes `= ` suffix where appropriate.

**Formatting is separate from parsing.** `FormatToml` in `formatting.go` uses its own regex-based line classifier; it does not call the loose parser.

**Position fidelity.** Any feature that depends on cursor location must go through `ParseTomlWithCursor`, not strict parse or plain string splits alone.

## Testing

| Package | Test file | Focus |
|---------|-----------|-------|
| `handlers/` | `diagnostics_test.go` | Type mismatches, required fields, unknown keys, AllOf |
| `handlers/` | `formatting_test.go` | Alignment, blank lines, comment preservation |
| `handlers/` | `folding_range_test.go` | Table/comment/string folding |
| `mappers/` | `schema_test.go` | Property map shape, `$ref` resolution |
| `models/` | `schema_lookup_test.go` | Comment-based schema detection |

Integration: open sample TOML in VS Code or Neovim with a known schema comment (`#helm`, etc.) and exercise completion, hover, diagnostics, format, and fold.

Manual checklist:

- Completion on `=`, space, and `#` (schema type picker)
- Hover on a key name inside a table
- Diagnostics for missing required field and unknown key
- Format document produces aligned keys within sections
- Folding on `[table]` sections
- Incomplete TOML does not crash the server

## Dependencies

Go:

- `github.com/tliron/glsp` — LSP protocol (3.16)
- `github.com/tliron/commonlog` — logging
- `github.com/invopop/jsonschema` — schema traversal
- `github.com/nuonco/nuon/pkg/config/schema` — embedded schemas
- `github.com/nuonco/nuon/pkg/parser/toml` — TOML parsing

VS Code extension:

- `vscode`, `vscode-languageclient` (Node; see `nuon-lsp-vscode/package.json`)

## Troubleshooting

**Server won't start:** confirm binary exists and is executable; stdio mode waits silently for input on stdin.

**No completions:** file must be TOML; first non-empty line must be a schema comment (`#typename`); cursor must be past the comment block for key/value completions.

**No diagnostics on unopened files:** workspace folders must be present at initialize (VS Code provides these; some editors only send `rootUri` — both paths are handled in `main.go`).

**Debug logging:** set `commonlog.Configure(0, nil)` in `main.go`; view VS Code Output panel "Nuon LSP" or Neovim `:LspLog`.

## References

- LSP spec: https://microsoft.github.io/language-server-protocol/
- glsp: https://github.com/tliron/glsp
- JSON Schema: https://json-schema.org/
- `README.md` — install and end-user docs
- `DESIGN.md` — parser and schema-mapping decisions
