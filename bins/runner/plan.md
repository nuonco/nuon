# Runner: fetch token as part of process

## Goals

1. avoid reading the token from disk or env
1. avoid requiring the token to be in the env. each process should fetch its own.
1. token per init for easier token refresh
1. pave the way for short-lived per task tokens

## Approach: Host Network + Self-Fetching Token

Both `runner mng` and `runner install` should fetch their own token on startup via IMDS. To make this work for
`runner install` (which runs in Docker), we switch the container to `--network host`. This gives the container direct
access to IMDS without hop limit issues.

The VM is not open to the internet in any way, so host networking has no meaningful security downside.

The fetch-token logic should be implemented as an fx provider that both processes share.

## `runner mng fetch-token` [DONE]

The `fetch-token` subcommand now supports a `--json` flag:

- **Without `--json`** (default): same behavior as before — authenticates via AWS IMDS, writes token to disk, prints
  human-readable output.
- **With `--json`**: authenticates via AWS IMDS, outputs JSON to stdout (including the token), does NOT write to disk.

Using the `Environment` directive — but note that `Environment` doesn't support shell expansion directly, so we use
`ExecStartPre` to write a drop-in, or use `PassEnvironment` from a wrapper. The cleanest approach is actually via
`ExecStartPre` + `systemctl set-environment`:

```ini
ExecStartPre=/bin/sh -c 'systemctl set-environment RUNNER_API_TOKEN=$(/usr/local/bin/runner mng fetch-token --json | jq -r .token)'
```

This sets `RUNNER_API_TOKEN` in the systemd manager's environment, making it available to `ExecStart` without any file
on disk.

### Implementation

- `FetchToken()` in `internal/jobs/management/fetch_token/standalone.go` — authenticates and returns result with token
  in memory (no disk write).
- `FetchAndStoreToken()` — wraps `FetchToken()`, writes to disk, clears token from result. Existing behavior preserved.
- `FetchTokenResult` now has json tags and an optional `Token` field (populated only by `FetchToken`).

## Host Network for Docker Container [DONE]

Switch the systemd service template from bridge networking to `--network host`:

1. Replace `-p 5000:5000` with `--network host` in `runner-service.aws.service`
2. Container binds directly to host ports — no port mapping needed
3. IMDS (`169.254.169.254`) is directly reachable — no hop limit issues

Since the VM is fully network-isolated (not open to the internet), there is no security downside to host networking. The
only consideration is port conflicts, but the VM is single-purpose (one container), so this is not a concern.

## Fetch Token as FX Provider [DONE]

The token fetch should happen during fx initialization, before the API client is constructed. Both `runner mng` and
`runner install` use `commonProviders()` which includes `api.New` — the API client provider that requires a token.

### Design

Create a new provider at `internal/pkg/auth/auth.go` (or similar) that:

1. Creates an unauthenticated runner SDK client (URL only, no token)
2. Calls `FetchToken()` to authenticate via IMDS
3. Returns a result that `api.New` can consume

The key constraint: the runner SDK client (`nuonrunner.Client`) bakes the token into the HTTP transport at creation
time. There is no `SetAuthToken()` method. So the token must be available _before_ `api.New` runs.

```
fx dependency chain:
  internal.NewConfig (loads runner_api_url, runner_id — but NOT token)
    -> auth.New (calls FetchToken via IMDS, returns token)
      -> api.New (constructs authenticated client using token from auth provider)
        -> settings.New (fetches settings using authenticated client)
          -> everything else
```

### Changes Required

1. **New provider `internal/pkg/auth/`** — fx provider that calls `FetchToken()` and returns the token (and runner ID,
   etc). This provider needs only the API URL (from config) to create an unauthenticated client for the auth call.

2. **Update `internal/config.go`** — make `runner_api_token` optional (remove `validate:"required"`). The token will
   come from the auth provider, not from env/config.

3. **Update `internal/pkg/api/api.go`** — accept the token from the auth provider instead of from config. Update
   `Params` to depend on the auth provider's output.

4. **Update `cmd/cli.go` `commonProviders()`** — add the auth provider to the chain, used by both `mng` and `install`.

5. **Update systemd service template** — switch to `--network host`, remove `EnvironmentFile=/opt/nuon/runner/token`,
   remove `--env-file /opt/nuon/runner/token`.

## Monitor Loop

The monitor loop runs as part of `runner mng`. This service runs directly on the VM. It is responsible for ensuring the
`runner install` systemd process is operational.

see: /Users/fd/nuon/runner/scripts/aws/init-mng.sh

With the new approach:

1. The monitor loop's own token is fetched at startup via the auth provider (same as `runner install`).
2. `ensureRunnerTokenFile` — remove entirely. No more token file on disk.
3. `ensureRunnerTokenValid` — still validates the token, but on failure re-fetches via IMDS and updates the in-memory
   client. This requires either reconstructing the client or adding a `SetAuthToken` method to the SDK.
4. The monitor no longer needs to manage the token file for the `runner install` process — each process fetches its own.

## `runner install`

This is the process that executes work. It runs in docker with `--network host`.

1. On startup, the auth fx provider fetches a token via IMDS (same code path as `mng`).
2. No token file on disk. No token in env. Fully self-sufficient.
3. The `EnvironmentFile=/opt/nuon/runner/token` and `--env-file /opt/nuon/runner/token` are removed from the service
   definition.

## TODO

- [x] Create `internal/pkg/auth/` provider that calls `FetchToken()` during fx init
- [x] Make `runner_api_token` config field optional
- [x] Update `api.New` to accept token from auth provider instead of config
- [x] Add auth provider to `commonProviders()` in `cmd/cli.go`
- [x] Update systemd service template: `--network host`, remove token file references
- [x] Remove `ensureRunnerTokenFile` from monitor loop (was already removed)
- [x] Update `ensureRunnerTokenValid` to use `FetchToken` (no disk write)
- [x] Remove `-p 5000:5000` from service template (not needed with host networking)
- [x] Add `SetAuthToken` to SDK client so monitor can refresh token without process restart
- [ ] Remove token file from `init-mng.sh` script

## Notes

not addressed but slated for change/removal:

1. here: https://github.com/nuonco/runner/blob/main/scripts/aws/init-mng.sh#L104-L108 we no longer _need_ to do this. We
   can leave it but the downstream processes are better suited for detecting failures which can be logged in
   cloudformation.
