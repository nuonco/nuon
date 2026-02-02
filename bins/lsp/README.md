# Nuon Language Server Protocol (LSP)

A Language Server Protocol implementation for Nuon TOML configuration files.

## Features

- **Code Completion**: Intelligent suggestions for configuration keys and values
- **Hover Information**: Documentation and type info on hover
- **Validation**: Real-time syntax checking

---

## Using the LSP

### VS Code

> **📹 Video walkthrough**: [Watch installation guide on Loom](https://www.loom.com/share/282efad7fd7c406d9c7b9286d2bcf98f)

#### Step 1: Check if you have the LSP binary

Open your terminal and run:
```bash
which nuon-lsp
```

✅ **If you see a path** (like `/usr/local/bin/nuon-lsp`) → Skip to Step 3
❌ **If you get "command not found"** → Continue to Step 2

#### Step 2: Install the LSP binary

The LSP binary comes with the Nuon CLI. Install it:

```bash
# Using Homebrew
brew install nuonco/tap/nuon

# OR using install script
curl -sSL install.nuon.co | bash
```

**Verify it worked:**
```bash
which nuon-lsp
```

You should see a path like: `/usr/local/bin/nuon-lsp`

#### Step 3: Install the VS Code extension

1. Open VS Code
2. Go to Extensions (Cmd/Ctrl+Shift+X)
3. Search for "Nuon LSP"
4. Click Install on the [Nuon LSP extension](https://marketplace.visualstudio.com/items?itemName=Nuon.nuon-lsp)

#### Step 4: Test it

1. Open any `.toml` file (or create a new one)
2. Add a schema type at the top:
   ```toml
   #helm
   ```
3. On a new line, type `[` - you should see completion suggestions

**That's it!** The extension automatically finds the `nuon-lsp` binary you installed.

#### Troubleshooting

**Extension not working?**

1. Check the LSP binary is in your PATH:
   ```bash
   which nuon-lsp
   ```

2. View error logs: Open VS Code Output panel (View → Output), select "Nuon LSP" from dropdown

3. Reload VS Code: Cmd/Ctrl+Shift+P → "Reload Window"

4. (Optional) Manually set the binary path in VS Code settings:
   ```json
   {
     "nuonLsp.serverPath": "/path/to/nuon-lsp"
   }
   ```

### Neovim

#### Step 1: Check if you have the LSP binary

Open your terminal and run:
```bash
which nuon-lsp
```

✅ **If you see a path** (like `/usr/local/bin/nuon-lsp`) → Skip to Step 3
❌ **If you get "command not found"** → Continue to Step 2

#### Step 2: Install the LSP binary

The LSP binary comes with the Nuon CLI. Install it:

```bash
# Using Homebrew
brew install nuonco/tap/nuon

# OR using install script
curl -sSL install.nuon.co | bash
```

**Verify it worked:**
```bash
which nuon-lsp
```

You should see a path like: `/usr/local/bin/nuon-lsp`

#### Step 3: Add to your Neovim config

Add this to your Neovim config file (usually `~/.config/nvim/init.lua`):

```lua
vim.lsp.start({
  name = "nuon-lsp",
  cmd = { "nuon-lsp" },
  root_dir = vim.fn.getcwd(),
  filetypes = { "toml" }
})
```

**Reload your config:**
```vim
:source $MYVIMRC
```

#### Step 4: Test it

1. Open any `.toml` file (or create a new one)
2. Check the LSP is attached:
   ```vim
   :LspInfo
   ```
   You should see "nuon-lsp" in the list

3. Try completions: Add a schema type at the top:
   ```toml
   # helm
   ```
   Then on a new line type `[` - you should see suggestions

**That's it!**

#### Troubleshooting

**LSP not attaching?**

1. Check the binary is in your PATH:
   ```bash
   which nuon-lsp
   ```

2. View LSP logs in Neovim:
   ```vim
   :LspLog
   ```

3. (Optional) Use absolute path in your config:
   ```lua
   vim.lsp.start({
     name = "nuon-lsp",
     cmd = { "/full/path/to/nuon-lsp" },  -- Use output from 'which nuon-lsp'
     root_dir = vim.fn.getcwd(),
     filetypes = { "toml" }
   })
   ```

---

## Development

The LSP server is a Go binary that communicates with editors via the Language Server Protocol.

### Architecture

The LSP has two runtime modes:

1. **Stdio Mode** (Production):
   - Used by editor extensions
   - Server communicates via stdin/stdout
   - Default mode when no flags are provided

2. **TCP Mode** (Development):
   - Server listens on a TCP port
   - Allows hot reloading without restarting your editor
   - Easier debugging with server logs in terminal

### Building the LSP Server

```bash
cd bins/lsp
go build -o nuon-lsp ./
```

This creates the `nuon-lsp` binary.

### Local Development with TCP Mode

For faster development iteration, you can run the LSP in TCP mode so you can restart it without reloading your editor.

#### Using nuonctl (Recommended)

```bash
nuonctl services dev --dev lsp
```

This automatically:
- Starts the LSP server on port 7001
- Starts health check on port 7002 (http://localhost:7002/health)
- Configures VS Code to use TCP port 7001
- Cleans up VS Code config on exit (Ctrl+C)

#### Manual Mode

```bash
cd bins/lsp
go run . -port 7001 -health-port 7002
```

Check health:
```bash
curl http://localhost:7002/health
```

Then manually configure VS Code settings:
```json
{
  "nuonLsp.port": 7001
}
```

**Benefits of TCP mode:**
- Restart the LSP server without reloading your editor
- See server logs directly in terminal
- Easier debugging with Go debugger
- Hot reload code changes

### Developing the VS Code Extension

The VS Code extension source is in `nuon-lsp-vscode/`.

> **Important**: You must open the `nuon-lsp-vscode` folder as your workspace root in VS Code. Opening a parent folder will prevent debugging from working correctly.

**Setup:**
1. Install dependencies:
   ```bash
   cd nuon-lsp-vscode && npm install
   ```

2. Build the LSP server:
   ```bash
   cd bins/lsp && go build -o nuon-lsp ./
   ```

3. Open `nuon-lsp-vscode` in VS Code:
   ```bash
   code nuon-lsp-vscode/
   ```

4. Press `F5` to open the Extension Development Host
   - The extension will activate automatically on startup
   - Check the Debug Console for: `🚀 Nuon-LSP extension activated`

5. Debug:
   - View server logs: Output panel → "Nuon LSP" dropdown
   - Debug client code: Use `console.log()` in `extension.js`

**Building for distribution:**
```bash
cd nuon-lsp-vscode
npm run vscode:prepublish
```

### Project Structure

```
bins/lsp/
├── main.go                    # LSP server entry point
├── handlers/                  # Request handlers (completion, hover, etc.)
├── mappers/                   # Schema mapping logic
├── models/                    # Data models
├── nuon-lsp-vscode/          # VS Code extension
│   ├── extension.js          # Extension client
│   └── package.json          # Extension metadata
├── README.md                  # This file
├── DESIGN.md                  # Architecture decisions
└── AGENTS.md                  # Developer guide for AI agents
```

### Adding New Features

**Adding a new LSP handler:**

1. Create handler in `handlers/`:
   ```go
   package handlers

   func TextDocumentNewFeature(ctx *glsp.Context, params *protocol.Params) (any, error) {
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

3. Update server capabilities in `initialize()` function in `main.go`

4. Rebuild: `go build -o nuon-lsp ./`

### Logging and Debugging

The server uses `commonlog` for structured logging. Adjust verbosity in `main.go`:

```go
commonlog.Configure(2, nil)  // 0=verbose, 2=minimal
```

**View logs:**
- **VS Code**: Output panel → "Nuon LSP"
- **Neovim**: `:LspLog` command

---

## Additional Notes

### Supported File Types

The LSP activates for `.toml` files. For the LSP to provide completions, your TOML file should include a schema type declaration as a comment at the top (usually added automatically by `nuon apps sync`):

```toml
# helm

[public_repo]
# ... configuration ...
```

The schema type is specified as a comment with `# <type>`. Supported types: `helm`, `docker_build`, `terraform_module`, `job`, and others.

### How It Works

The LSP uses:
- **Custom TOML parser** - Handles incomplete input gracefully (essential while typing)
- **JSON Schema mapping** - Provides context-aware completions based on your configuration type
- **Position tracking** - Knows whether you're editing a key or a value to suggest appropriate completions

See [DESIGN.md](DESIGN.md) for detailed architecture information.

## References

- [Language Server Protocol Specification](https://microsoft.github.io/language-server-protocol/)
- [Nuon Documentation](https://docs.nuon.co)
- [DESIGN.md](DESIGN.md) - Architecture decisions
- [AGENTS.md](AGENTS.md) - Developer guide for AI assistants
