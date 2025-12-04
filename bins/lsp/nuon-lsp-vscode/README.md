# Nuon LSP Extension (Internal)

VS Code extension for Nuon Language Server.

## Installation

1. Install LSP binary:
   ```bash
   # Via install script (installs both CLI and LSP)
   curl -sSL install.nuon.co | bash

   # Or via Homebrew (installs both CLI and LSP)
   brew install nuonco/tap/nuon
   ```

2. Install extension:
   ```bash
   cd bins/lsp/nuon-lsp-vscode
   npm install
   npm run build
   npm run install
   ```

## Development

Open this folder in VS Code and press F5 to debug.

## Configuration

Custom server path (optional):
```json
{
  "nuonLsp.serverPath": "/custom/path/to/nuon-lsp"
}
```
