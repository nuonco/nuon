# Nuon Language Server Protocol (LSP)

A Language Server Protocol implementation for Nuon configuration files, providing code completion, hover information, and validation for TOML. 

## Features

- **Code Completion**: Intelligent suggestions with trigger characters (`=` and space)
- **Hover Information**: Contextual documentation on hover
- **Text Document Sync**: Full document synchronization

## Building the LSP

```bash
cd bins/lsp
go build -o nuon-lsp ./
```

This creates the `lsp` binary that serves as the language server.

For VSCode extension, also copy the binary to the extension's `bin` directory:

```bash
mkdir -p nuon-lsp-vscode/bin
cp lsp nuon-lsp-vscode/bin/
```

## Development Setup

The LSP runs as a subprocess communicating via stdio. Configuration differs between editors.

### Neovim

#### Configuration

Add to your Neovim config:

```lua
vim.lsp.start({
  name = "nuon-lsp",
  cmd = { "/path/to/mono/bins/lsp/lsp" },  -- Update with your actual path
  root_dir = vim.fn.getcwd()
})
```

Replace `/path/to/mono` with your actual repository path.

#### Development Workflow

1. **Rebuild the binary** after changes:
   ```bash
   cd bins/lsp && go build -o nuon-lsp ./
   ```

2. **Restart Neovim** or reload your config to pick up the new binary:
   ```vim
   :source $MYVIMRC
   ```

3. **Check LSP status**:
   ```vim
   :LspInfo
   ```

4. **View server logs**:
   ```vim
   :edit /tmp/nvim-lsp-debug.log  " Or check Neovim's LSP log
   ```

### VSCode

#### Configuration

The VSCode extension is located in `nuon-lsp-vscode/`.

The extension configuration is in `nuon-lsp-vscode/extension.js`:

```javascript
// Server path is configurable via nuonLsp.serverPath setting
const config = vscode.workspace.getConfiguration("nuonLsp");
let serverCommand = config.get("serverPath");
if (!serverCommand || serverCommand === "") {
    serverCommand = path.join(context.extensionPath, "..", "lsp");
}

const serverOptions = {
  command: serverCommand,
  args: [],
  transport: TransportKind.stdio,
};

const clientOptions = {
  documentSelector: [
    { scheme: "file", language: "toml" },
  ],
};
```

#### Installation

1. **Build and prepare the extension**:
   ```bash
   cd bins/lsp
   go build -o nuon-lsp ./
   mkdir -p nuon-lsp-vscode/bin
   cp lsp nuon-lsp-vscode/bin/
   cd nuon-lsp-vscode && npm install
   ```

2. **Install the extension in VSCode**:
   - Copy the extension folder to your VSCode extensions directory:
     ```bash
     cp -r nuon-lsp-vscode ~/.vscode/extensions/nuon.nuon-lsp-0.0.1
     ```
   - Install dependencies in the installed extension:
     ```bash
     cd ~/.vscode/extensions/nuon.nuon-lsp-0.0.1 && npm install
     ```
   - Reload VSCode

#### Development

> **Important**: You must open the `nuon-lsp-vscode` folder as your workspace root in VSCode. Opening a parent folder will prevent debugging from working correctly.

1. **Install Node dependencies**:
   ```bash
   cd nuon-lsp-vscode && npm install
   ```

2. **Update the binary path** in `extension.js` if needed

3. **Run the extension in development mode**:
   - Open VSCode with `nuon-lsp-vscode` as the workspace root: `code nuon-lsp-vscode/`
   - Press `F5` to open the Extension Development Host
   - The extension will activate automatically on startup

4. **Check the Debug Console** for activation logs:
   - Look for: `🚀 Nuon-LSP extension activated`
   - The output will show connection status

#### Debugging the Extension

1. In the Extension Development Host, open the VSCode Output panel (`View > Output`)
2. Select "Nuon LSP" from the dropdown to see server logs
3. Use `console.log()` in `extension.js` for debugging client-side code

#### Building for Distribution

```bash
cd nuon-lsp-vscode
npm run vscode:prepublish
```

## Project Structure

- `main.go` - Entry point and LSP initialization
- `handlers/` - Request handlers (completion, hover, did-change, etc.)
- `mappers/` - Data structure mappings
- `models/` - Data models
- `nvim.lua` - Neovim configuration example
- `nuon-lsp-vscode/` - VSCode extension
  - `extension.js` - Extension activation and client setup
  - `package.json` - Extension metadata and dependencies
- `schema.json` - Configuration schema for validation
- `example.toml` - Example Nuon configuration file

## Supported File Types

- `.toml` - Nuon application manifests

> **Note**: For the LSP to provide completions, your TOML file must declare a schema type at the top (before any `[table]` headers):
> ```toml
>  # helm
> ```
> Supported types include: `helm`, `docker_build`, `terraform_module`, `job`, and others defined in the Nuon JSON schema.

## Adding New Features

### Adding a New Handler

1. Create a handler function in `handlers/`
2. Register it in `main.go`:
   ```go
   handler = protocol.Handler{
     TextDocumentNewFeature: handlers.TextDocumentNewFeature,
   }
   ```
3. Rebuild: `go build -o nuon-lsp ./`

### Modifying Language Support

Update the `clientOptions.documentSelector` in VSCode's `extension.js` and the `activationEvents` in `package.json`.

## Troubleshooting

### LSP won't start in Neovim

- Verify the binary path is correct: `which lsp` or the full path you specified
- Check Neovim LSP logs: `:LspLog` or check `/tmp/nvim-*.log`
- Ensure the binary is executable: `chmod +x bins/lsp/lsp`

### VSCode extension doesn't activate

- Check the Debug Console for errors
- Verify the binary path in `extension.js` is correct
- Ensure the binary is executable: `chmod +x bins/lsp/lsp`
- Restart the Extension Development Host (`Ctrl+Shift+F5`)

### No completions showing up

- Ensure the file type matches the `documentSelector` in the editor config
- Check that the LSP server is running: `:LspInfo` (Neovim) or Output panel (VSCode)
- Verify trigger characters are recognized: `=` and space

## Logging

The server uses `commonlog` for logging. Adjust verbosity in `main.go`:

```go
commonlog.Configure(2, nil)  // Change verbosity level here
```

## References

- [Language Server Protocol Specification](https://microsoft.github.io/language-server-protocol/)
- [glsp - Go LSP Framework](https://github.com/tliron/glsp)
- [VSCode Language Client](https://code.visualstudio.com/docs/extensions/language-support)
