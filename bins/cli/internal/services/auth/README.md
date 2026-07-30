# Auth Service

The auth service handles CLI authentication via `nuon auth login` and `nuon auth logout`.

## Login Flow

### 1. API URL Selection

The login flow starts by determining which API endpoint to authenticate against.

**If no API URL is configured** (first-time user, no `api_url` in `~/.nuon` or `NUON_API_URL` env):

```
? Which Nuon deployment are you using?
  > Nuon Cloud
    Nuon BYOC
```

- **Nuon Cloud** proceeds with `https://api.nuon.co`.
- **Nuon BYOC** prompts for a custom API URL.

**If an API URL is already configured** (returning user), shows the URL with its source and a confirmation:

```
  https://api.nuon.co (~/.nuon)
? Login to Nuon Cloud (Y/n)
```

or for custom URLs:

```
  https://api.custom-domain.com (NUON_API_URL env)
? Login to https://api.custom-domain.com (Y/n)
```

The source label is tracked by `Config.APIURLSource` which is either the config file path (e.g. `~/.nuon`) or
`NUON_API_URL env`.

Declining the confirmation prompts for a new URL.

### 2. Authentication

The CLI fetches auth configuration from the API (`GetCLIConfig`) and gets a device code flow against the Nuon auth
service. To do this, it generates a local device code, opens a browser to `auth.<root_domain>/device/code`, and polls
for a token.

### 3. Org Selection

After authentication, the CLI checks the user's org memberships:

- **0 orgs**: Prompts user to create one or request an invite.
- **1 org**: Auto-selects it and saves to config.
- **Multiple orgs**: Prompts user to select one.

### 4. Config Persistence

The selected API URL, access token, and org ID are saved to `~/.nuon`.

## Logout

`nuon logout` clears `api_token` and `api_url` from the config file. The next `nuon login` will show the deployment type
selector since no URL is configured.

## Configuration Sources

The API URL is resolved by viper in priority order:

1. `NUON_API_URL` environment variable
2. `api_url` in `~/.nuon` config file
3. Struct default: `https://api.nuon.co` (not visible to viper, used only when no explicit value is set)

## OIDC Workload Identity Federation (CI)

For automation, the CLI can exchange an ambient OIDC ID token for a short-lived Nuon API token — no
stored secrets. The org must have a matching OIDC trust policy (`nuon orgs oidc-trust-policies`).

Two entry points:

- **Transparent**: when no `api_token` is configured, an org is selected (`NUON_ORG_ID` or config),
  and an ambient OIDC source is available, `doPersistentPreRunE` exchanges automatically. The token
  is kept in-memory only — every invocation exchanges fresh, so expiry never needs handling.
- **Explicit**: `nuon auth exchange-token` prints the exchanged token (table output prints only the
  token, for `export NUON_API_TOKEN=$(nuon auth exchange-token)`).

Ambient token sources, in precedence order (`internal/oidctoken`):

1. `--oidc-token` flag (explicit command only)
2. `NUON_OIDC_TOKEN` — a raw OIDC JWT
3. `NUON_OIDC_TOKEN_FILE` — path to a file containing the JWT
4. GitHub Actions — `ACTIONS_ID_TOKEN_REQUEST_URL`/`ACTIONS_ID_TOKEN_REQUEST_TOKEN` (requires
   `permissions: id-token: write`); the audience is set from `--audience` / `NUON_OIDC_AUDIENCE`

The exchange calls `POST /v1/oidc/token` (unauthenticated) with `{org_id, token}`; the API verifies
the JWT against the org's trust policies (issuer JWKS, audience, claim conditions) and mints a
short-lived token bound to the policy's service account.
