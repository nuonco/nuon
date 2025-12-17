# Nuon LSP Extension

Language Server Protocol support for Nuon TOML configuration files in Visual Studio Code.

## Features

- **Syntax Validation**: Real-time validation of Nuon TOML configuration files
- **Auto-completion**: Intelligent suggestions for Nuon configuration options
- **Hover Documentation**: View documentation for configuration keys on hover
- **Go to Definition**: Navigate to configuration definitions
- **Error Diagnostics**: Clear error messages and warnings for configuration issues

## Installation

### Prerequisites

1. Install the Nuon LSP server binary:
   ```bash
   # Via install script (installs both CLI and LSP)
   curl -sSL install.nuon.co | bash

   # Or via Homebrew (installs both CLI and LSP)
   brew install nuonco/tap/nuon
   ```

### Install Extension

#### From VSIX (Recommended)

```bash
cd bins/lsp/nuon-lsp-vscode
npm install
npm run build
npm run install
```

#### From VS Code Marketplace

*(Coming soon)*

## Usage

Once installed, the extension automatically activates when you open `.toml` files. The LSP server will:

- Validate your Nuon configuration syntax
- Provide completion suggestions
- Show documentation on hover
- Display error diagnostics

## Configuration

### Custom Server Path

If you have a custom build or different installation path for the Nuon LSP server:

```json
{
  "nuonLsp.serverPath": "/custom/path/to/nuon-lsp"
}
```

Leave empty to use the default (looks for `nuon-lsp` in your PATH or bundled binary).

### Development Mode (TCP)

For local LSP server development, you can connect to a running server via TCP:

```json
{
  "nuonLsp.port": 7001
}
```

Set to `0` (default) to use stdio mode with the local binary.

## Development

### Building from Source

```bash
cd bins/lsp/nuon-lsp-vscode
npm install
npm run build
```

### Debugging

1. Open this folder in VS Code
2. Press F5 to launch the Extension Development Host
3. Open a `.toml` file to test the extension

### Running LSP Server Locally

For development with a local LSP server:

```bash
# Terminal 1: Run LSP server in TCP mode
cd bins/lsp
go run main.go --port 7001

# Terminal 2: Configure VS Code to connect
# Set "nuonLsp.port": 7001 in settings
```

## Troubleshooting

### Extension not activating

- Ensure you have `.toml` files open
- Check the "Nuon LSP" output channel in VS Code for errors
- Verify `nuon-lsp` binary is in your PATH: `which nuon-lsp`

### LSP server not found

If you see "Nuon LSP server not found" errors:

1. Install the Nuon CLI/LSP: `brew install nuonco/tap/nuon`
2. Verify installation: `nuon-lsp --version`
3. Set custom path in settings if needed

### Connection issues

- Check the VS Code Output panel (View > Output > Nuon LSP)
- Ensure no other process is using the configured port
- Try restarting VS Code

## Support

- **Issues**: [GitHub Issues](https://github.com/nuonco/nuon/issues)
- **Documentation**: [Nuon Docs](https://docs.nuon.co)
- **Website**: [nuon.co](https://nuon.co)

## License

MIT
