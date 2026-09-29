const path = require("path");
const vscode = require("vscode");
const fs = require("fs");
const net = require("net");
const { execSync } = require("child_process");
const { LanguageClient, TransportKind } =
  require("vscode-languageclient/node");

let client;

function findInPath(binaryName) {
  try {
    const command = process.platform === "win32" ? "where" : "which";
    const result = execSync(`${command} ${binaryName}`, { encoding: "utf8" });
    return result.trim().split("\n")[0];
  } catch {
    return null;
  }
}

function activate(context) {
  console.log("🚀 Nuon-LSP extension activated");

  const config = vscode.workspace.getConfiguration("nuonLsp");
  const port = config.get("port") || 0;

  if (port > 0) {
    console.log(`📍 Connecting to LSP dev server on localhost:${port}`);

    const serverOptions = () => {
      return new Promise((resolve, reject) => {
        const socket = net.connect({ port: port, host: "127.0.0.1" });
        socket.on("connect", () => {
          console.log("✅ TCP socket connected");
          resolve({
            reader: socket,
            writer: socket
          });
        });
        socket.on("error", (err) => {
          console.error("❌ TCP socket error:", err);
          reject(err);
        });
      });
    };

    const clientOptions = {
      documentSelector: [
        { scheme: "file", language: "toml" },
      ],
      synchronize: {
        fileEvents: vscode.workspace.createFileSystemWatcher("**/*.toml"),
      },
      outputChannel: vscode.window.createOutputChannel("Nuon LSP"),
    };

    client = new LanguageClient(
      "nuonLsp",
      "Nuon LSP",
      serverOptions,
      clientOptions
    );

    const clientStartPromise = client.start();
    context.subscriptions.push({
      dispose: () => client.stop()
    });

    clientStartPromise
      .then(() => {
        console.log("✅ Nuon-LSP client ready (TCP mode on port " + port + ")");
        vscode.window.showInformationMessage("Nuon-LSP connected (dev mode)");
      })
      .catch(err => {
        console.error("❌ Nuon-LSP TCP connection failed:", err);
        vscode.window.showErrorMessage(`Failed to connect to LSP server on port ${port}: ${err.message}`);
      });

    return;
  }

  startStdioMode(context);
}

function startStdioMode(context) {
  const config = vscode.workspace.getConfiguration("nuonLsp");
  let serverCommand = config.get("serverPath");

  if (!serverCommand || serverCommand === "") {
    const pathBinary = findInPath("nuon-lsp");
    if (pathBinary) {
      serverCommand = pathBinary;
      console.log(`📍 Found nuon-lsp in PATH: ${serverCommand}`);
    } else {
      serverCommand = path.join(context.extensionPath, "..", "lsp");
    }
  }

  if (!fs.existsSync(serverCommand)) {
    const errorMsg = `Nuon LSP server not found at: ${serverCommand}`;
    console.error("❌", errorMsg);
    vscode.window.showErrorMessage(errorMsg);
    return;
  }

  console.log(`📍 Using LSP server at: ${serverCommand}`);

  const serverOptions = {
    command: serverCommand,
    args: [],
    options: {
      stdio: 'pipe'
    }
  };

  const clientOptions = {
    documentSelector: [
      { scheme: "file", language: "toml" },
    ],
    synchronize: {
      fileEvents: vscode.workspace.createFileSystemWatcher("**/*.toml"),
    },
    outputChannel: vscode.window.createOutputChannel("Nuon LSP"),
  };

  console.log("⚙️  Starting Nuon-LSP client");
  client = new LanguageClient(
    "nuonLsp",
    "Nuon LSP",
    serverOptions,
    clientOptions
  );

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
