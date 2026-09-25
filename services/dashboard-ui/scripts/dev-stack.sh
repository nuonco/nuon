#!/usr/bin/env bash
set -eu

usage() {
  cat <<'EOF'
Start the dashboard BFF, SPA live-reload proxy, and Ladle in this checkout.

Usage:
  bun run dev:stack -- [port]
  ./scripts/dev-stack.sh [port]
  HTTP_PORT=4010 bun run dev:stack

Arguments:
  port            Go BFF port (proxy is port+1). Default: 4000 or $HTTP_PORT

Options:
  --ladle-port N  Ladle port (default: 61000 + (port - 4000), next free if taken)
  -h, --help      Show this help

Ctrl-C stops everything. If a process exits, the others keep running and a
status block is reprinted.
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DASHBOARD_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
cd "$DASHBOARD_DIR"

PORT="${HTTP_PORT:-}"
LADLE_PORT="${LADLE_PORT:-}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    -h|--help)
      usage
      exit 0
      ;;
    --ladle-port)
      LADLE_PORT="${2:-}"
      if [[ -z "$LADLE_PORT" ]]; then
        echo "error: --ladle-port requires a value" >&2
        exit 1
      fi
      shift 2
      ;;
    --)
      shift
      ;;
    -*)
      echo "error: unknown option $1" >&2
      usage >&2
      exit 1
      ;;
    *)
      PORT="$1"
      shift
      ;;
  esac
done

PORT="${PORT:-4000}"
if ! [[ "$PORT" =~ ^[0-9]+$ ]]; then
  echo "error: port must be a number (got '$PORT')" >&2
  exit 1
fi
PROXY_PORT=$((PORT + 1))

if ! command -v go >/dev/null 2>&1; then
  echo "error: go is not on PATH" >&2
  exit 1
fi
if ! command -v bun >/dev/null 2>&1; then
  echo "error: bun is not on PATH" >&2
  exit 1
fi
if [[ ! -d "$DASHBOARD_DIR/node_modules" ]]; then
  echo "error: node_modules missing; run bun install in $DASHBOARD_DIR" >&2
  exit 1
fi

if [[ -t 1 ]]; then
  C_BFF='\033[36m'
  C_SPA='\033[32m'
  C_LADLE='\033[35m'
  C_ERR='\033[1;31m'
  C_OK='\033[1;32m'
  C_DIM='\033[2m'
  C_RESET='\033[0m'
else
  C_BFF=''; C_SPA=''; C_LADLE=''; C_ERR=''; C_OK=''; C_DIM=''; C_RESET=''
fi

port_in_use() {
  local p="$1"
  if command -v ss >/dev/null 2>&1; then
    [[ -n "$(ss -ltnH "sport = :${p}" 2>/dev/null)" ]]
    return
  fi
  bash -c "echo >/dev/tcp/127.0.0.1/${p}" >/dev/null 2>&1 && return 0
  bash -c "echo >/dev/tcp/::1/${p}" >/dev/null 2>&1
}

wait_port() {
  local p="$1" label="$2" tries="${3:-150}"
  local i
  for i in $(seq 1 "$tries"); do
    if port_in_use "$p"; then
      return 0
    fi
    sleep 0.1
  done
  printf '%berror:%b timed out waiting for %s on :%s\n' "$C_ERR" "$C_RESET" "$label" "$p" >&2
  return 1
}

pick_free_port() {
  local p="$1"
  while port_in_use "$p"; do
    p=$((p + 1))
  done
  echo "$p"
}

href() {
  local u="$1"
  if [[ -t 1 ]]; then
    printf '\033]8;;%s\033\\%s\033]8;;\033\\' "$u" "$u"
  else
    printf '%s' "$u"
  fi
}

if [[ -z "$LADLE_PORT" ]]; then
  LADLE_PORT=$((61000 + PORT - 4000))
  if [[ "$LADLE_PORT" -lt 1 ]]; then
    LADLE_PORT=61000
  fi
  LADLE_PORT="$(pick_free_port "$LADLE_PORT")"
elif ! [[ "$LADLE_PORT" =~ ^[0-9]+$ ]]; then
  echo "error: ladle port must be a number (got '$LADLE_PORT')" >&2
  exit 1
fi

if [[ ! -x "$DASHBOARD_DIR/node_modules/.bin/ladle" ]]; then
  echo "error: ladle is not installed; run bun install in $DASHBOARD_DIR" >&2
  exit 1
fi

STATE_KEY="$(printf '%s' "$DASHBOARD_DIR" | sha1sum | awk '{print substr($1,1,12)}')"
STATE_DIR="${TMPDIR:-/tmp}/nuon-dashboard-stack-${STATE_KEY}"
mkdir -p "$STATE_DIR"
PIDS_FILE="$STATE_DIR/pids"
SUPERVISOR_FILE="$STATE_DIR/supervisor"
BIN="$STATE_DIR/server"
LADLE_BIN="$DASHBOARD_DIR/node_modules/.bin/ladle"

stop_group() {
  local pid="$1"
  if [[ -z "$pid" ]]; then
    return
  fi
  kill -TERM -- "-$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null || true
  sleep 0.15
  kill -KILL -- "-$pid" 2>/dev/null || kill -KILL "$pid" 2>/dev/null || true
}

if [[ -f "$SUPERVISOR_FILE" ]]; then
  OLD_SUP="$(tr -d '[:space:]' < "$SUPERVISOR_FILE" || true)"
  if [[ -n "${OLD_SUP:-}" && "$OLD_SUP" != "$$" ]] && kill -0 "$OLD_SUP" 2>/dev/null; then
    kill -TERM "$OLD_SUP" 2>/dev/null || true
    for _ in $(seq 1 30); do
      kill -0 "$OLD_SUP" 2>/dev/null || break
      sleep 0.1
    done
    kill -KILL "$OLD_SUP" 2>/dev/null || true
  fi
fi
if [[ -f "$PIDS_FILE" ]]; then
  # shellcheck disable=SC2046
  for old in $(cat "$PIDS_FILE"); do
    stop_group "$old"
  done
  rm -f "$PIDS_FILE"
fi
echo "$$" > "$SUPERVISOR_FILE"

prefix() {
  local name="$1" color="$2"
  while IFS= read -r line || [[ -n "$line" ]]; do
    printf '%b[%s]%b %s\n' "$color" "$name" "$C_RESET" "$line"
  done
}

BFF_STATUS=starting
SPA_STATUS=starting
LADLE_STATUS=starting
crashed=0

print_status() {
  echo
  printf '%b── dashboard stack ────────────────────────────────%b\n' "$C_DIM" "$C_RESET"
  printf '  bff     '
  href "http://localhost:${PORT}"
  printf '   %s\n' "$BFF_STATUS"
  printf '  proxy   '
  href "http://localhost:${PROXY_PORT}"
  printf '   %s\n' "$SPA_STATUS"
  printf '  ladle   '
  href "http://localhost:${LADLE_PORT}"
  printf '   %s\n' "$LADLE_STATUS"
  printf '%b──────────────────────────────────────────────────%b\n' "$C_DIM" "$C_RESET"
  echo
}

mark_crash() {
  local name="$1" code="$2"
  crashed=1
  printf '\n%b● %s exited (status %s)%b\n' "$C_ERR" "$name" "$code" "$C_RESET"
  case "$name" in
    bff) BFF_STATUS="crashed (status $code)" ;;
    spa) SPA_STATUS="crashed (status $code)" ;;
    ladle) LADLE_STATUS="crashed (status $code)" ;;
  esac
  print_status
}

alive() {
  local pid="${1:-}"
  [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null
}

BFF_PID=
SPA_PID=
LADLE_PID=
CLEANED=0
STOPPING=0

write_pids() {
  printf '%s %s %s\n' "${BFF_PID:-}" "${SPA_PID:-}" "${LADLE_PID:-}" > "$PIDS_FILE"
}

cleanup() {
  if [[ "$CLEANED" -eq 1 ]]; then
    return
  fi
  CLEANED=1
  STOPPING=1
  trap - EXIT INT TERM
  echo
  printf '%bStopping dashboard stack…%b\n' "$C_DIM" "$C_RESET"
  stop_group "${LADLE_PID:-}"
  stop_group "${SPA_PID:-}"
  stop_group "${BFF_PID:-}"
  rm -f "$PIDS_FILE" "$SUPERVISOR_FILE"
}

trap cleanup EXIT
trap 'cleanup; exit 130' INT
trap 'cleanup; exit 143' TERM

printf 'Building Go BFF from %s…\n' "${NUON_DIR:-$REPO_DIR}"
go build -C "${NUON_DIR:-$REPO_DIR}" -o "$BIN" ./services/dashboard-ui/server

export HTTP_PORT="$PORT"

set -m 2>/dev/null || true

rm -f dist/.port
"$BIN" serve > >(prefix bff "$C_BFF") 2>&1 &
BFF_PID=$!

for _ in $(seq 1 50); do
  [[ -f dist/.port ]] && break
  sleep 0.1
done

if [[ -f dist/.port ]]; then
  ACTUAL="$(tr -d '[:space:]' < dist/.port)"
  if [[ -n "$ACTUAL" && "$ACTUAL" != "$PORT" ]]; then
    printf '%bnote:%b BFF bound :%s instead of :%s (port in use)\n' "$C_DIM" "$C_RESET" "$ACTUAL" "$PORT"
    PORT="$ACTUAL"
    PROXY_PORT=$((PORT + 1))
    export HTTP_PORT="$PORT"
  fi
fi

if ! wait_port "$PORT" "bff"; then
  BFF_STATUS="failed to bind"
  print_status
  exit 1
fi
BFF_STATUS="running"

bun run dev > >(prefix spa "$C_SPA") 2>&1 &
SPA_PID=$!

"$LADLE_BIN" dev --port "$LADLE_PORT" > >(prefix ladle "$C_LADLE") 2>&1 &
LADLE_PID=$!
write_pids
set +m 2>/dev/null || true

SPA_UP=0
LADLE_UP=0
for _ in $(seq 1 300); do
  if [[ "$SPA_UP" -eq 0 ]] && port_in_use "$PROXY_PORT"; then
    SPA_UP=1
    SPA_STATUS="running"
  fi
  if [[ "$LADLE_UP" -eq 0 ]] && port_in_use "$LADLE_PORT"; then
    LADLE_UP=1
    LADLE_STATUS="running"
  fi
  if [[ "$SPA_UP" -eq 1 && "$LADLE_UP" -eq 1 ]]; then
    break
  fi
  sleep 0.1
done
if [[ "$SPA_UP" -eq 0 ]]; then
  SPA_STATUS="not listening"
  printf '%berror:%b timed out waiting for proxy on :%s\n' "$C_ERR" "$C_RESET" "$PROXY_PORT" >&2
fi
if [[ "$LADLE_UP" -eq 0 ]]; then
  LADLE_STATUS="not listening"
  printf '%berror:%b timed out waiting for ladle on :%s\n' "$C_ERR" "$C_RESET" "$LADLE_PORT" >&2
fi

print_status
if [[ "$SPA_UP" -eq 1 && "$LADLE_UP" -eq 1 ]]; then
  printf '%bAll three are up.%b Click a URL above to open it.\n\n' "$C_OK" "$C_RESET"
fi

set +e
while [[ "$STOPPING" -eq 0 && ( -n "$BFF_PID" || -n "$SPA_PID" || -n "$LADLE_PID" ) ]]; do
  changed=0
  if [[ "$STOPPING" -eq 1 ]]; then
    break
  fi
  if [[ -n "$BFF_PID" ]] && ! alive "$BFF_PID"; then
    wait "$BFF_PID"
    mark_crash bff "$?"
    BFF_PID=
    changed=1
  fi
  if [[ -n "$SPA_PID" ]] && ! alive "$SPA_PID"; then
    wait "$SPA_PID"
    mark_crash spa "$?"
    SPA_PID=
    changed=1
  fi
  if [[ -n "$LADLE_PID" ]] && ! alive "$LADLE_PID"; then
    wait "$LADLE_PID"
    mark_crash ladle "$?"
    LADLE_PID=
    changed=1
  fi
  write_pids
  if [[ "$changed" -eq 0 ]]; then
    sleep 0.5
  fi
done

if [[ "$crashed" -eq 1 ]]; then
  exit 1
fi
exit 0
