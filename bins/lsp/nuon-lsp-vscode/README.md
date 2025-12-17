# Nuon LSP Extension

Language Server Protocol support for Nuon TOML configuration files in Visual Studio Code.

## Features in Action

**Auto-completion for Nuon configuration:**

![Auto-completion](https://i.postimg.cc/50Jky0SM/Screenshot-2025-12-17-at-8-29-27-PM.png)

**Hover documentation and type information:**

![Hover info](https://i.postimg.cc/tTbmJTNy/Screenshot-2025-12-17-at-8-29-43-PM.png)

## Quick Start

1. Install the Nuon CLI (includes the LSP server):
   ```bash
   brew install nuonco/tap/nuon
   ```
   Or use the install script: `curl -sSL install.nuon.co | bash`

2. Install the Nuon LSP extension from the [marketplace](https://marketplace.visualstudio.com/items?itemName=Nuon.nuon-lsp).

3. Open a Nuon TOML configuration file (`.toml`) and the extension will activate automatically.

**Note:** The Nuon CLI must be installed for the extension to work. See the [CLI installation guide](https://nuoninc.mintlify.app/cli) for detailed instructions.

## Features

- **Syntax Validation** - Real-time validation of Nuon TOML configuration files
- **Auto-completion** - Intelligent suggestions for Nuon configuration options
- **Hover Documentation** - View documentation for configuration keys on hover
- **Go to Definition** - Navigate to configuration definitions
- **Error Diagnostics** - Clear error messages and warnings for configuration issues

## Settings

The extension works out of the box with no configuration required. It automatically detects the `nuon-lsp` binary from your Nuon CLI installation.

**Optional settings:**
- `nuonLsp.serverPath` - Custom path to the nuon-lsp binary (leave empty to auto-detect)

## Troubleshooting

If the extension is not working:

1. Verify the Nuon CLI is installed: `nuon-lsp --version`
2. Check the "Nuon LSP" output channel in VS Code for error details
3. Ensure `nuon-lsp` is in your PATH: `which nuon-lsp`
4. Try reloading VS Code: Cmd/Ctrl+Shift+P → "Reload Window"

## Resources

- [Nuon Documentation](https://docs.nuon.co/get-started/introduction)
- [CLI Installation Guide](https://docs.nuon.co/cli)
- [GitHub Issues](https://github.com/nuonco/nuon/issues)
- [Website](https://nuon.co)
