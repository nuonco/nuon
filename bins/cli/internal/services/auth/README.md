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

The CLI fetches auth configuration from the API (`GetCLIConfig`), then uses one of two flows:

- **Nuon Auth** (`NuonAuthEnabled: true`): Device code flow against the Nuon auth service. Generates a local device
  code, opens a browser to `auth.<root_domain>/device/code`, and polls for a token.
- **Auth0** (`NuonAuthEnabled: false`): Standard Auth0 device code flow with OAuth2 token exchange.

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
