# Auth Service

This is technically a full stack app. We are building it to conform to the way we have built the other HTTP services,
which are all APIs, to maintain our patterns. At a future date, we will refactor our fx deps and the `cmd`s. For now, we
respect convention.

**In Scope**

- Allow BYOC Nuon customers to BYO-IdP.

**Out of Scope**

- Allow Cloud orgs to BYO-IdP. This is a future feature, not a today feature.

**Plan for**

- Allowing users to connect their account to multiple IdP providers.

## Prior Art

This service borrows heavily from https://github.com/vouch/vouch-proxy.

## Configs

New env vars for auth.nuon.co

| Env var               | Description                                             | Scope        |
| --------------------- | ------------------------------------------------------- | ------------ |
| ROOT_DOMAIN           | The root domain for all the services. e.g. byoc.org.co. | All Services |
| NUON_AUTH_SESSION_KEY | Session Key for nonce                                   | Auth Service |
| NUON_AUTH_JWT_SECRET  | Secret for siging the jwt                               | Auth Service |

We construct a domain from the root domain by prefixing `auth` to it instead of using a NUON_AUTH_DOMAIN env var.

## Identity Providers

This service depends on a the table `identity_providers` and on the default identity provider configured via env vars.

### Configs

There are two sources for configurations.

1. Env vars: a default provider must be provided via env vars. If these configs are not provided, the service will not
   start up.
2. Database: additionaly `IdentityProvider`s can be created. These store the configs in a json column.

#### Env Var Configs

| Env var                 | Description                       |
| ----------------------- | --------------------------------- |
| nuon_auth_provider_type | one of 'oidc', 'google', 'github' |
| nuon_auth_client_id     | Client ID                         |
| nuon_auth_client_secret | Client Secret                     |
| nuon_auth_issuer_url    | Issuer URL                        |
| nuon_auth_redirect_url  | Redirect URL                      |

#### Examples

For these examples, assume the BYOC Nuon is deployed with the following ROOT_DOMAIN: `byoc.org.co`.

**Auth0** as a generic OIDC provider

| Env var                 | Value                        |
| ----------------------- | ---------------------------- |
| nuon_auth_provider_type | 'oidc'                       |
| nuon_auth_client_id     | `[secret]`                   |
| nuon_auth_client_secret | `[secret]`                   |
| nuon_auth_issuer_url    | `https://audience.auth0.com` |
| nuon_auth_redirect_url  | `auth.byoc.org.co/auth`      |

**Google** as an OAuth provider

| Env var                 | Value                   |
| ----------------------- | ----------------------- |
| nuon_auth_provider_type | 'google'                |
| nuon_auth_client_id     | `[secret]`              |
| nuon_auth_client_secret | `[secret]`              |
| nuon_auth_redirect_url  | `auth.byoc.org.co/auth` |

### in-database configs

If additional IdPs must be created, we need to add to the `identity_providers` table. At the time of writing, only
install-wide configs are supported. This means any providers added this way will be available to all users of the
install. As such, these are created via and admin api endpoint.

Future: in the future we'll add support for org-specific IdPs. At that point, an org-specific endpont on the public api
may be added.

### Examples

**Auth0** as a generic OIDC provider

```json
{
  "enabled": true,
  "openid_config": {
    "client_id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "client_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "issuer_url": "https://audience.auth0.com",
    "redirect_url": "http://auth.byoc.org.co/auth"
  },
  "provider_type": "oidc"
}
```

### Loading Order

The default provider is required for startup. this default provider is composed from the provider configs loaded from
the env files.

After startup, we load the rest of the providers with a helper called `getIdentityProviders`. This helper loads the
default provider from the env and then loads any additional providers from the database. If a provider config is
malformed, we omit it from the list and return only the valid providers. This should be avoided by validating provider
configs otw in, though.

Providers are identified by their ProviderType. This is used in the UI and in the urls to determine which provider to
load and use.

### Provider Lookup

| Function                                       | Purpose                                                         |
| ---------------------------------------------- | --------------------------------------------------------------- |
| `getProviderByType(ctx, app.ProviderType)`     | Main entry point - returns configured `Provider` by type enum   |
| `getIdentityProviderByType(ctx, ProviderType)` | Looks up `IdentityProvider` by type (checks env first, then DB) |
| `getDefaultIdentityProvider()`                 | Builds `IdentityProvider` from env vars only                    |
| `createProviderFromIdentityProvider(ip)`       | Converts `IdentityProvider` model to `Provider` interface       |

### Uniqueness

At the time of writing, only one of each type (google, github, oidc) provider can be provided. This limitation includes
the provider configured via env vars.

## Internals

### Templates

This is the only service that is using templates. Even then, all templates are "namespaced" with auth so they are all
registered with this naming convention `auth/*.tmpl`. The files are all named `*.html.tmpl`.

### Middleware

This service does need or know about the Org or Auth middleware in use by the other API services. This service has only
a few endpoints so it doesn't actually imlement an auth middleware for its own use.

### Cookie

This service uses a cookie `X-Nuon-Auth` for auth. The cookie is a JWT with the following keys:

| attr  | desc |
| ----- | ---- |
| token |      |
| email |      |

## Forthcoming

#### SAML

SAML support will be built into this service.

#### SCIM

SCIM support will be built into this service.

### Additional Claims

We'd like to add support for additional claims so the IdP can provide additional details about the user. This is a
future-looking feature leaving room for the concept of roles and org access.

| attr   | desc                                                                    |
| ------ | ----------------------------------------------------------------------- |
| groups | list of strings where each string is a group such as admin or developer |
| orgs   | list of strings where each string is, ideally, an org id or an org slug |

## TODO (now)

- [x] use NuonRootDomain instead of NuonAuthDomain
- [x] Add `identity_provider` model (org_id + provider type are unique together) (consider default provider type as
      well).
- [x] Drop JWT in favor of simpler cookie w/ token
- [x] Add support for allowed_domains (email domains) or domain limitation.

## \*Caveats

- Changing an env-var provider to a different provider of the same type is likely to cause issues. At this time, this is
  not supported.
- Changing the provider, site-wide or for an org, of a given type to a different provider of the same type will cause
  issues. At this point, this is unsupported until we develop a story for handling changes in `sub`. Theoretically,
  email is enough to preserve the connectio to the account.
- Users are not able to arbitrarily sign up for the service. Successful auth requires for the user to have an account or
  an org invite.

### Agents

The following requirements were confirmed during implementation:

1. **Provider Implementation**: Use [vouch-proxy](https://github.com/vouch/vouch-proxy/) as inspiration for implementing
   Google, GitHub, and OpenID Connect providers in `internal/pkg/auth/providers`.

2. **Session Management**: Use custom signed cookies (HMAC-SHA256) instead of gin-contrib/sessions. Set `SameSite=None`
   because the auth cookie will be used across multiple domains.

3. **Identity Provider Configuration**:

   - Provider-specific config structs (`OpenIDConfig`, `GoogleConfig`, `GitHubConfig`) with validation based on
     `ProviderType`
   - Store configs in a JSONB column, `configs`, on the `IdentityProvider` model

4. **Provider Loading Sources**:

   - **Environment variables**: Default provider (required). Service will not start without valid config.
   - **Database**: Additional `IdentityProvider` records (optional, loaded after migrations are applied)

5. **Service Startup Validation**: Use service-level validation pattern - the `New()` function returns
   `(*service, error)` and validates the default provider config. If required env vars (`nuon_auth_provider_type`,
   `nuon_auth_client_id`, `nuon_auth_client_secret`, `nuon_auth_redirect_url`) are not provided or invalid, the service
   fails to start via FX.
