# Auth Service

This is technically a full stack app. We are building it to conform to the way we have built the other HTTP services,
which are all APIs, to maintain our patterns. At a future date, we will refactor our fx deps and the `cmd`s. For now, we
respect convention.

## Prior Art

This service borrows heavily from https://github.com/vouch/vouch-proxy.

## Configs

New env vars for auth.nuon.co

| Env var                   | Description                                                                                 | Scope        |
| ------------------------- | ------------------------------------------------------------------------------------------- | ------------ |
| ROOT_DOMAIN               | The root domain for all the services. e.g. byoc.org.co.                                     | All Services |
| NUON_AUTH_SESSION_KEY     | Session Key for nonce                                                                       | Auth Service |
| NUON_AUTH_ALLOWED_DOMAINS | Domains from which any user can register. Usually the org email.                            | Auth Service |
| NUON_AUTH_ALLOW_ALL_USERS | Whether or not any user with an email from an allowed domain should be allowed to register. | Auth Service |

We construct a domain from the root domain by prefixing `auth` to it. e.g. `app`.`byoc.org.co`. This becomes the service
domain which the service understands itself to be available at .

<!-- prettier-ignore-start -->
> [!IMPORTANT]
> These values are required. If they do are not present, the service will not start.

> [!NOTE]
> `NUON_AUTH_ALLOW_ALL_USERS` defaults to false.

> [!NOTE]
> `NUON_AUTH_ALLOWED_DOMAINS` cannot be empty.

<!-- prettier-ignore-end -->

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
  "provider_type": "oidc",
  "openid_config": {
    "client_id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "client_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
    "issuer_url": "https://audience.auth0.com",
    "redirect_url": "http://auth.byoc.org.co/auth"
  }
}
```

**Github**

```json
{
  "enabled": true,
  "provider_type": "github",
  "github_config": {
    "client_id": "XXXXXXXXXXXXXXXXXXXX",
    "allowed_orgs": ["999999999"],
    "redirect_url": "https://auth.byoc.org.co/auth",
    "allowed_teams": ["1111111"],
    "client_secret": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  }
}
```

Default scopes are, `user:email` and `read:user`.

<!-- prettier-ignore-start -->
> [!IMPORTANT]
> If a user does not have a public email, they will not be able to authenticate.

> [!IMPORTANT]
> If a user's primary email does not have an allowed domain, they will not be able to authenticate.

<!-- prettier-ignore-end -->

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
the provider configured via env vars. What this means is that if an oidc provider is configured for this service via env
vars, only `google` or `github` oidc providers can be added via in-database configs.

## Internals

### Templates

This is the only service that is using templates. Even then, all templates are "namespaced" with auth so they are all
registered with this naming convention `auth/*.tmpl`. The files are all named `*.html.tmpl`.

### Middleware

This service does need or know about the Org or Auth middleware in use by the other API services. This service has only
a few endpoints so it doesn't actually implement an auth middleware for its own use and instead relies on the session.

### Session

The auth service maintains a short lived session, `nuon-auth-session`, to store the next url (final destination) and the
provider type. This session is deleted as soon as the auth flow is complete and is only ever available to the auth
service itself. The service uses the `NUON_AUTH_SESSION_KEY` to sign the cookie for the session (HMAC-SHA256).

### Cookie

When auth is complete, the auth provider sets a cross domain cookie `X-Nuon-Auth` that is accessible to any service on
`ROOT_DOMAIN` (e.g. `byoc.org.co`). This cookie is invalidated when a user visits `/logout`.

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

## Caveats & Limitations

### Changing a Provider of a Given Type

Changing an env-var provider to a different provider of the same type is likely to cause issues. At this time, this is
not explicitly supported. This applies to both, env-var based IdPs and in-database IdPs.
