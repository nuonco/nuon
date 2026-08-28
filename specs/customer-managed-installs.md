# Customer-managed installs: online portal

## Product contract

The online customer portal uses the standard Nuon API authentication contract. An administrator supplies the portal
with these values out of band:

- ctl-api URL;
- organization ID;
- install ID;
- a standard token minted for the install's dedicated portal service account.

Secret storage and token configuration are outside this milestone. The browser never receives the API token. The
portal backend sends every ctl-api request with:

```http
Authorization: Bearer <service-account-token>
X-Nuon-Org-ID: <org-id>
```

## Identity and authorization

Each customer-managed install has a dedicated standard service account. It is bound to a portal-specific role whose
policy contains only `permissions.Object(orgID, permissions.KindInstall, installID)`. It receives no organization-wide
role and does not share the stack service account role.

Portal endpoints are normal authenticated public routes. Each route declares its install resource and permission with
`internal/pkg/authz/require`:

- release, package, workflow, log, and approval reads require `read`;
- workflow retry requires `update`;
- package grants and approval responses require `create`.

Feature-flag and operating-model checks remain in the handlers. Release discovery, release files, packages, and
download grants are constrained to the install's app and configured app config.

## Install creation

`POST /v1/customer-managed/installs` creates an ordinary customer-managed install and starts its normal provisioning
workflow. It also creates the `InstallOperatingModel`, ensures the dedicated portal service account,
and binds its install-scoped role. The response includes the install, operating model, and non-secret service-account metadata
so an administrator can mint and configure a token separately.

There is no setup invitation, enrollment secret, custom token format, connection record, portal heartbeat, health
credential lifecycle, or custom authentication middleware.

## Data model

The online milestone keeps:

- `InstallOperatingModel`, which records connectivity and customer/vendor authority;
- `InstallReleaseDeployment`, which records the release outcome associated with normal ctl-api workflows.

It adds no portal authentication or connection tables. Standard `Account`, `Role`, `Policy`, `AccountRole`, and `Token`
records are the source of truth for portal identity and authentication.

## Portal configuration

Connected portal flags are:

```text
--control-plane-url
--org-id
--install-id
--api-token
```

Normal ctl-api workflows, plans, logs, and approvals remain the source of truth. Any disconnected operating mode is a
separate follow-up and must not replace or bypass this authentication and authorization contract.

## Release source capture

Customer-managed app configs preserve the exact TOML consumed by the config parser and local text files resolved by
`features:"get"` fields. The parser returns this source alongside the semantic app config and records explicit release
member identities, so source capture does not scan arbitrary paths or reinterpret TOML. CLI syncs upload that source,
and VCS-triggered app branch runs capture it from their checked-out repository. The existing Temporal blob data
converter externalizes large activity payloads to S3.

The VCS-triggered release path requires end-to-end verification with a real app branch before this surface ships.
