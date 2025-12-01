const path = require("path");
const vscode = require("vscode");
const fs = require("fs");
const { LanguageClient, TransportKind } =
  require("vscode-languageclient/node");

let client;

function activate(context) {
  console.log("🚀 Nuon-LSP extension activated");
  vscode.window.showInformationMessage("Nuon-LSP extension activated!");

  const config = vscode.workspace.getConfiguration("nuonLsp");
  let serverCommand = config.get("serverPath");

  if (!serverCommand || serverCommand === "") {
    serverCommand = path.join(context.extensionPath, "..", "lsp");
  }

  // Validate that the server binary exists
  if (!fs.existsSync(serverCommand)) {
    const errorMsg = `Nuon LSP server binary not found at: ${serverCommand}`;
    console.error("❌", errorMsg);
    vscode.window.showErrorMessage(errorMsg);
    return;
  }

  console.log(`📍 Using LSP server at: ${serverCommand}`);

  const serverOptions = {
    command: serverCommand,
    args: [], // or ["--stdio"] if your server expects it
    transport: TransportKind.stdio,
  };

  const clientOptions = {
    documentSelector: [
      { scheme: "file", language: "toml" },
    ],
  };

  console.log("⚙️  Starting Nuon-LSP client");
  client = new LanguageClient(
    "nuonLsp",               // internal ID
    "Nuon LSP",              // display name (appears in Output)
    serverOptions,
    clientOptions
  );

  // Start the client and register with subscriptions
  const clientStartPromise = client.start();
  context.subscriptions.push({
    dispose: () => client.stop()
  });

  clientStartPromise
    .then(() => {
      console.log("✅ Nuon-LSP client ready");
    })
    .catch(err => {
      console.error("❌ Nuon-LSP failed:", err);
      vscode.window.showErrorMessage(`Nuon-LSP failed to start: ${err.message}`);
    });

}

function deactivate() {
  if (client) {
    return client.stop();
  }
}

module.exports = { activate, deactivate };
