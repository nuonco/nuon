#!/usr/bin/env bash
# Runs the mintlify dev server for local docs development.
#
# mint refuses to run on node 25, and bun reports itself as node 25+, so `npx`
# backed by bun (or an active node 25) fails. This finds a supported node and
# runs mint under it.
set -euo pipefail

MINT_VERSION="${MINT_VERSION:-4.2.314}"
PORT="${PORT:-3333}"

# node is supported if it is real node (not bun) and >= 20.17, != 25.x
node_supported() {
  local bin="$1"
  [[ -x "$bin" ]] || return 1
  "$bin" -e '
    if (typeof Bun !== "undefined" || typeof Deno !== "undefined") process.exit(1);
    const [maj, min] = process.versions.node.split(".").map(Number);
    process.exit(maj === 25 || maj < 20 || (maj === 20 && min < 17) ? 1 : 0);
  ' >/dev/null 2>&1
}

find_node() {
  local candidates=()
  [[ -n "${MINT_NODE:-}" ]] && candidates+=("$MINT_NODE")
  candidates+=("$(command -v node || true)")
  for v in 24 22 20; do
    candidates+=("/opt/homebrew/opt/node@$v/bin/node" "/usr/local/opt/node@$v/bin/node")
    while IFS= read -r n; do candidates+=("$n"); done < <(
      ls -d "$HOME"/.nvm/versions/node/v$v.*/bin/node 2>/dev/null | sort -rV
    )
  done

  for c in "${candidates[@]}"; do
    if node_supported "$c"; then
      echo "$c"
      return 0
    fi
  done
  return 1
}

if ! NODE_BIN="$(find_node)"; then
  echo "docs: no supported node found (need >= 20.17, not 25.x, not bun)." >&2
  echo "      install one (e.g. 'brew install node@24') or set MINT_NODE=/path/to/node" >&2
  exit 1
fi

NODE_DIR="$(dirname "$NODE_BIN")"
echo "docs: using node $("$NODE_BIN" --version) from $NODE_DIR"

# docs/.npmrc sets min-release-age, which makes npx refuse the pinned mint version
export NPM_CONFIG_MIN_RELEASE_AGE=0
export PATH="$NODE_DIR:$PATH"

exec "$NODE_DIR/npx" --yes "mint@${MINT_VERSION}" dev --no-open --port "$PORT"
