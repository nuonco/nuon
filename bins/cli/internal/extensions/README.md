# CLI Extensions System

Internal developer documentation for the `nuon ext` extension system.

For user-facing documentation, see [docs/guides/cli-extensions.mdx](/docs/guides/cli-extensions.mdx).

## Overview

The extension system allows third-party and first-party commands to be installed and run as top-level `nuon`
subcommands. It is inspired by the
[GitHub CLI extension model](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions).

The feature is gated behind `NUON_PREVIEW=true`. When preview mode is off:

- The `nuon ext` command group is hidden and errors on use.
- No extension proxy commands are registered on the root command.

Debug logging is available via `NUON_DEBUG=true`.

## Package Layout

```
bins/cli/internal/extensions/
├── types.go      # Core types: ExtType, ExtensionManifest, InstalledExtension
├── manager.go    # Manager: List, Get, Remove, BinaryPath, EnsureDir
├── install.go    # Install, InstallLocal, normalizeRepo, release/clone logic, archive extraction
├── upgrade.go    # Upgrade, UpgradeAll
├── exec.go       # Exec: dispatch to binary/script/python
├── browse.go     # Browse: search nuonco org for extensions with the nuon-extensions topic
└── manifest.go   # FetchManifest, ParseManifest, ValidateManifest, CheckCLIVersion

bins/cli/cmd/
├── extensions.go # Cobra commands: list, install, upgrade, remove, browse, exec
└── root.go       # Proxy command registration (extensionProxyCmd)
```

## Extension Types

Type detection is automatic. The CLI never asks the user what type an extension is.

| Type     | Constant        | Detection                                       | Execution                |
| -------- | --------------- | ----------------------------------------------- | ------------------------ |
| `binary` | `ExtTypeBinary` | GitHub Release with platform asset, or fallback | Direct binary exec       |
| `script` | `ExtTypeScript` | Executable `nuon-ext-<name>` at repo root       | Direct script exec       |
| `python` | `ExtTypePython` | `pyproject.toml` at repo root                   | `uv run nuon-ext-<name>` |

Detection order in `detectExtType`:

1. `pyproject.toml` exists → `python`
2. `nuon-ext-<name>` executable exists at root → `script`
3. Otherwise → `binary`

## Install Flow

### Entry point: `Manager.Install(repo string)`

`Install` routes to one of three paths based on the input:

```
input starts with . / ~ ?
  → InstallLocal (symlink)

input has @ref suffix?
  → try getReleaseByTag(repo, ref)
    → platform asset found? → installByRelease (download binary from that tag's release)
    → no asset?             → installByClone (git clone at that ref)

no ref?
  → try getLatestRelease(repo)
    → platform asset found? → installByRelease (download binary from latest release)
    → no asset?             → installByClone (git clone default branch)
```

### Name Resolution (`normalizeRepo`)

The `@ref` suffix is stripped first, then the input is resolved:

| Input                           | repo                  | name  | ref         |
| ------------------------------- | --------------------- | ----- | ----------- |
| `api`                           | `nuonco/nuon-ext-api` | `api` | `""`        |
| `nuon-ext-api`                  | `nuonco/nuon-ext-api` | `api` | `""`        |
| `nuonco/nuon-ext-api`           | `nuonco/nuon-ext-api` | `api` | `""`        |
| `myorg/nuon-ext-foo`            | `myorg/nuon-ext-foo`  | `foo` | `""`        |
| `nuonco/nuon-ext-api@v0.19.798` | `nuonco/nuon-ext-api` | `api` | `v0.19.798` |
| `api@main`                      | `nuonco/nuon-ext-api` | `api` | `main`      |

Shorthand names (no `/`) always resolve to the `nuonco` org. The repo name **must** use the `nuon-ext-` prefix.

### Release Asset Matching (`findReleaseAsset`)

The expected asset base name is:

```
nuon-ext-<name>-<GOOS>-<GOARCH>[.exe]
```

The function checks for matches in this order:

1. Bare binary (e.g. `nuon-ext-api-darwin-arm64`)
2. `.tar.gz` archive (e.g. `nuon-ext-api-darwin-arm64.tar.gz`)
3. `.zip` archive (e.g. `nuon-ext-api-darwin-arm64.zip`)

**Gotcha**: GoReleaser produces `.tar.gz` archives by default, not bare binaries. The matching must check all three
formats. This was the root cause of a bug where `nuon ext install nuonco/nuon-ext-api` silently fell through to
`installByClone` because the bare binary name didn't match the `.tar.gz` asset name.

### Archive Extraction (`downloadAndExtractBinary`)

When the download URL ends in `.tar.gz`, the binary is streamed and extracted from the archive. The extractor matches by
`filepath.Base(hdr.Name)` so it handles archives that contain entries like `./nuon-ext-api` or just `nuon-ext-api`.

For bare binary downloads, the file is saved directly and `chmod 0755` is applied.

### Version Pinning with `@ref`

When a `@ref` is provided:

1. The manifest is fetched from that ref (so `min_cli_version` is validated against the pinned code).
2. The CLI tries `getReleaseByTag(repo, ref)` to see if a GitHub Release exists for that tag.
3. If a release with a matching platform asset is found, it downloads the binary — same as latest.
4. If no release exists (e.g. the ref is a branch name or commit SHA), it falls back to `installByClone`.

This means `nuon ext install api@v0.19.798` downloads the compiled binary from the `v0.19.798` release, while
`nuon ext install api@main` clones the repo at the `main` branch.

### Local Install (`InstallLocal`)

Local install creates a **symlink** from `~/.config/nuon/extensions/nuon-ext-<name>` → the source directory. This means:

- Rebuilding the binary in the source dir takes effect immediately.
- Editing scripts takes effect immediately.
- `manifest.json` is written into the source directory (through the symlink).
- `nuon ext remove <name>` removes only the symlink, not the source.

Requirements:

- The directory must be named `nuon-ext-<name>` (matching the manifest's `extension.name`).
- The directory must contain a `nuon-ext.toml`.
- For binary extensions, the compiled binary must already exist in the directory.

## Upgrade Flow

`Upgrade` only works for compiled extensions installed from a release:

1. Fetches `getLatestRelease` for the extension's repo.
2. Compares `release.TagName` against the installed `Tag`.
3. If different, downloads the new binary (with archive extraction) and updates `manifest.json`.

`UpgradeAll` iterates all installed extensions and calls `Upgrade` on each.

Interpreted extensions (script/python) installed via clone do not have an upgrade path — they should be removed and
re-installed.

## Execution (`Exec`)

The `Exec` method dispatches based on `ext.Type`:

| Type     | Command                 | Working Directory |
| -------- | ----------------------- | ----------------- |
| `binary` | `<extDir>/<binary>`     | inherited         |
| `script` | `<extDir>/<entrypoint>` | `<extDir>`        |
| `python` | `uv run <entrypoint>`   | `<extDir>`        |

All extensions receive:

- Full inherited environment (`os.Environ()`)
- CLI context variables: `NUON_API_URL`, `NUON_ORG_ID`, `NUON_APP_ID`, `NUON_INSTALL_ID`, `NUON_API_TOKEN`,
  `NUON_CONFIG_FILE`
- Extension-specific: `NUON_EXT_NAME`, `NUON_EXT_DIR`

Auth warnings are printed to stderr if `requires_token` or `requires_org` are set but the values are missing. The
extension still runs — auth enforcement is the extension's responsibility.

Extensions get stdin/stdout/stderr passthrough and the CLI exits with the extension's exit code.

## Proxy Commands

When `NUON_PREVIEW=true`, `root.go` registers each installed extension as a top-level cobra command via
`extensionProxyCmd`. This makes `nuon api` equivalent to `nuon ext exec api`.

If an extension name collides with a built-in command (e.g. `auth`, `config`), the built-in always wins. The
`reservedCommandNames` map in `extensions.go` tracks these names, and a warning is printed at install time. The user can
still run the extension via `nuon ext exec <name>`.

## Browse

`Browse` searches the `nuonco` GitHub org for repos matching `nuon-ext-*` that have the `nuon-extensions` topic. Only
repos with this topic appear in browse results. Third-party extensions can be installed directly but won't show up in
browse.

## Storage Layout

```
~/.config/nuon/extensions/
├── nuon-ext-api/                          # compiled (release download)
│   ├── nuon-ext-api                       # platform binary (extracted from .tar.gz)
│   ├── nuon-ext.toml                      # cached manifest from repo
│   └── manifest.json                      # install metadata
├── nuon-ext-gen-readme/                   # interpreted (git clone)
│   ├── pyproject.toml
│   ├── nuon-ext.toml
│   ├── manifest.json
│   └── src/
├── nuon-ext-my-tool -> /home/user/...     # local dev (symlink)
```

`manifest.json` is the source of truth for the extension manager. It contains:

- `name`, `description`, `repo`, `version`, `tag`, `ref`
- `type` (`binary`, `script`, `python`)
- `binary` (filename, only for binary type)
- `entrypoint` (filename, only for script/python type)
- `platform` (`darwin/arm64`, etc.)
- `requires_token`, `requires_org`

## GitHub API Usage

All GitHub API calls are unauthenticated. This means:

- Rate limits are 60 requests/hour per IP.
- Only public repositories are supported.
- The `User-Agent` header is set to `nuon-cli/<version>`.

API endpoints used:

- `GET /repos/{owner}/{repo}/releases/latest` — latest release
- `GET /repos/{owner}/{repo}/releases/tags/{tag}` — release by tag
- `GET /repos/{owner}/{repo}/contents/nuon-ext.toml` — manifest via contents API (supports `?ref=`)
- `GET /search/repositories?q=...` — browse
- `https://raw.githubusercontent.com/{owner}/{repo}/{ref}/nuon-ext.toml` — raw manifest for caching

## Development Workflow

### Working on the extension system itself

```bash
# Build the CLI
cd bins/cli
go build -o nuon-dev .

# Test with preview mode and debug logging
NUON_PREVIEW=true NUON_DEBUG=true ./nuon-dev ext install nuonco/nuon-ext-api
NUON_PREVIEW=true NUON_DEBUG=true ./nuon-dev ext install nuonco/nuon-ext-api@v0.19.798

# Clean up between test runs
rm -rf ~/.config/nuon/extensions/nuon-ext-api

# List installed extensions
NUON_PREVIEW=true ./nuon-dev ext list

# Test execution
NUON_PREVIEW=true ./nuon-dev ext exec api --help

# Test upgrade
NUON_PREVIEW=true NUON_DEBUG=true ./nuon-dev ext upgrade api
```

### Building an extension locally

```bash
# Clone or create the extension repo
cd ~/nuon/nuon-ext-my-tool

# Build the binary (for compiled extensions)
go build -o nuon-ext-my-tool .

# Install from local directory (creates symlink)
NUON_PREVIEW=true nuon ext install ./nuon-ext-my-tool

# The symlink means rebuilds take effect immediately
go build -o nuon-ext-my-tool .
nuon my-tool --help   # picks up the new binary

# When done developing, remove and install from GitHub
nuon ext remove my-tool
nuon ext install myorg/nuon-ext-my-tool
```

### Verifying release assets

When debugging install failures, check what the release actually contains:

```bash
# List releases
gh release list --repo nuonco/nuon-ext-api

# Inspect asset names (this is what findReleaseAsset matches against)
gh release view v0.19.798 --repo nuonco/nuon-ext-api --json assets --jq '.assets[].name'

# Download and inspect an archive
gh release download v0.19.798 --repo nuonco/nuon-ext-api \
  --pattern "nuon-ext-api-darwin-arm64.tar.gz" -D /tmp
tar tzf /tmp/nuon-ext-api-darwin-arm64.tar.gz
```

## Known Gotchas

1. **GoReleaser produces `.tar.gz`, not bare binaries.** `findReleaseAsset` must check for `.tar.gz` and `.zip` suffixes
   in addition to the bare binary name. If you add support for a new archive format, update both `findReleaseAsset`
   (matching) and `downloadAndExtractBinary` (extraction).

2. **`Upgrade` shares the same asset-matching logic.** If you change `findReleaseAsset` or archive extraction in
   `install.go`, make sure `upgrade.go` stays in sync — it calls `findReleaseAsset` and `downloadAndExtractBinary`
   directly.

3. **No GitHub authentication.** All API calls are unauthenticated. Private extension repos won't work. The 60 req/hour
   rate limit can be hit during development — if you get 403s, wait or add a `GITHUB_TOKEN` header (not currently
   implemented).

4. **`manifest.json` lands in the source directory for local installs.** Because local install creates a symlink and
   then writes `manifest.json` through it, the file ends up in the extension's source tree. You may want to `.gitignore`
   it in extension repos.

5. **`detectExtType` has a subtle precedence issue.** If a Python project also has an executable `nuon-ext-<name>`
   script at the root, the script detection takes precedence over Python. This is because the Python check looks for
   `pyproject.toml` but then falls through if the script file also exists. This is intentional — it lets Python projects
   provide a wrapper script.

6. **Clone fallback for SHA refs.** `git clone --depth 1 --branch <ref>` works for branches and tags but not commit
   SHAs. For SHAs, the code does a full (non-shallow) clone followed by `git checkout <ref>`. This is slower but
   necessary.

7. **Reserved command names.** The `reservedCommandNames` map in `cmd/extensions.go` must be kept in sync with the
   actual top-level commands registered in `root.go`. If a new built-in command is added, add its name to the map.
