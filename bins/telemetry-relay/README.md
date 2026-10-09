# Nuon BYOC Telemetry Relay

An OpenTelemetry Collector distribution that forwards logs, metrics, and traces from Nuon install runners,
in-cluster collectors, and the environment hosting the relay to an OTLP/HTTP backend. Both ingestion paths
are enabled by default.

| Path | Default listener | Processing |
|---|---|---|
| Install | `0.0.0.0:4318` | Verify runner or collector JWT, check optional org allowlist, replace identity IDs with verified values, apply install resource attributes |
| Environment | `0.0.0.0:5318` | No authentication or identity validation; preserve incoming attributes except configured resource changes |

Both paths accept `/v1/logs`, `/v1/metrics`, and `/v1/traces`. The environment path supports application,
infrastructure, and control-plane telemetry without requiring a Nuon control-plane or downstream install ID.
The relay also sends its own metrics through this path.

The install listener requires a Nuon control plane or compatible token issuer; arbitrary JWTs are not supported.
Signing keys must load successfully before either listener starts.

Forwarding is synchronous, with no relay queue or retry loop; senders own buffering and retries.

## Configuration

The supplied [Collector configuration](config.yaml) uses these environment variables:

| Variable | Purpose |
|---|---|
| `NUON_TELEMETRY_ISSUER` | Expected JWT issuer |
| `NUON_TELEMETRY_AUDIENCE` | Required JWT audience: the exact public relay endpoint configured in the control plane, whether deployment-default or org-specific. |
| `NUON_TELEMETRY_ALLOW_LEGACY_AUDIENCE` | Also accept runner tokens with the legacy audience `urn:nuon:telemetry`; defaults to `true`. Collector tokens always require endpoint binding. Set to `false` to require it for runners too. |
| `NUON_TELEMETRY_JWKS_URL` | Issuer's public signing-key URL |
| `NUON_TELEMETRY_JWKS_ALLOW_INSECURE` | Permit HTTP issuer and JWKS URLs; defaults to `false` |
| `NUON_TELEMETRY_ALLOWED_ORG_IDS` | YAML/JSON list of allowed JWT org IDs for install telemetry; unset or `[]` allows all verified orgs |
| `NUON_TELEMETRY_ENVIRONMENT_ENDPOINT` | Environment receiver bind address; defaults to `0.0.0.0:5318` |
| `NUON_TELEMETRY_SELF_OTLP_ENDPOINT` | Relay self-metrics OTLP/HTTP endpoint; defaults to `http://127.0.0.1:5318` |
| `VENDOR_OTLP_ENDPOINT` | Backend OTLP/HTTP base URL |
| `VENDOR_OTLP_AUTHORIZATION` | Backend `Authorization` header value |

Issuer and JWKS URLs require HTTPS by default. Enable HTTP only over a trusted network; exact JWT issuer matching
and HTTPS certificate verification remain enforced. Keep private signing keys on the issuer.

Terminate TLS in front of the install listener. The environment listener is intended for operator-managed internal
access: do not route public ingress to it. The relay does not enforce this networking policy. For local-only
producers, set `NUON_TELEMETRY_ENVIRONMENT_ENDPOINT=127.0.0.1:5318`. If changing the port or binding an address that
does not accept loopback traffic, update `NUON_TELEMETRY_SELF_OTLP_ENDPOINT` to reach that listener too.

Health checks listen on `0.0.0.0:13133`; keep them private. Prometheus self-metrics remain available on
`127.0.0.1:8888/metrics`.

## Org-specific relay endpoints

An org admin can set `relay_endpoint` on `PATCH /v1/orgs/current/telemetry` to an HTTPS OTLP/HTTP base URL.
Setting it to `null` or an empty string restores the control plane's deployment-default relay; omitting it preserves the current
endpoint. The `enabled` setting is independent and can be omitted when changing only the endpoint.

For a customer-run relay, set `NUON_TELEMETRY_AUDIENCE` to the exact registered endpoint (including any path or
trailing slash), configure the control plane's issuer and JWKS URL, and set `NUON_TELEMETRY_ALLOWED_ORG_IDS` to
that org's ID. Backend credentials stay on the relay. For new customer-run relays that do not need legacy-token
compatibility, set `NUON_TELEMETRY_ALLOW_LEGACY_AUDIENCE=false`. Keep the unauthenticated environment listener private.

By default, the relay accepts either its configured endpoint or `urn:nuon:telemetry` as a token's single audience,
supporting both older and newer runners. Legacy-audience acceptance has no automatic cutoff.
Signature, issuer, expiry, telemetry scope, identity checks, and any org allowlist still apply. Legacy tokens
are not bound to one destination: any relay trusting the issuer, accepting that audience, and permitting the
token's org can accept them.

Runner settings resolve org overrides for all runners. Newer runners include the selected `relay_endpoint`
when requesting a token; the control plane verifies it matches current settings and uses it as the audience.
Older runners omit this parameter and receive `urn:nuon:telemetry` tokens, including on renewal. Keep legacy-audience
acceptance enabled on any relay serving those runners.

On an endpoint change, updated runners stop the old exporter and token renewal before obtaining credentials
for the new endpoint. A stale token request is rejected; the runner refreshes settings and retries. A failed
switch does not resume sending to the old destination. Existing disk-queued telemetry may be forwarded to the
new destination; endpoint changes are not a queue-drain or lossless-cutover mechanism.

## In-cluster collector authentication

A collector uses two credentials:

- An opaque Nuon API token for the install's managed `telemetry_collector` service account, used to fetch
  settings and request relay credentials from the control plane's runner API.
- A short-lived JWT, used to send telemetry to the relay's install listener.

Authenticate runner API requests with `Authorization: Bearer <Nuon API token>`:

1. Poll `GET /v1/installs/:install_id/telemetry/collector-settings`. The response contains `enabled`,
   `relay_endpoint`, and `resource_attributes`. Stop forwarding when `enabled` is false, but continue polling
   so the collector can detect re-enablement.
2. When enabled, request credentials with
   `POST /v1/installs/:install_id/telemetry/collector-access-token?relay_endpoint=<URL-encoded relay_endpoint>`.
   Supply the exact endpoint returned by settings. A stale endpoint returns HTTP 409; refresh settings before retrying.
3. Use the returned `access_token` in `Authorization: Bearer <access_token>` for OTLP requests to the relay.
   The response includes `token_type: "Bearer"` and `expires_in: 600`; renew before expiry.

Token issuance requires create permission on the install's telemetry resource and a live managed-account binding
for that install and collector purpose. A runner or org admin cannot obtain a collector JWT using their own identity.
Telemetry must be enabled and the relay configured. Settings require read permission and remain accessible while
forwarding is disabled.

Collector JWTs are RS256-signed, use `typ=at+jwt`, carry only `telemetry:write`, and have a ten-minute lifetime.
Their single audience is the exact relay endpoint. The subject is
`org:<org_id>:install:<install_id>:collector:<account_id>`; `client_id` and `nuon_collector_id` identify the
managed account, and `nuon_runner_id` is absent. Org, app, and install IDs are resolved by the control plane.
Collectors cannot use legacy-audience tokens, even when the relay accepts those tokens for runners.

The relay verifies JWTs without consulting the control plane. Disabling telemetry, revoking the Nuon API token,
or removing the managed account prevents new JWTs but does not invalidate those already issued. They remain valid
until expiry, including the relay's 30-second clock-skew allowance. The relay must support collector identities;
a relay that only accepts runner identities rejects collector JWTs.

## Install identity and organization filtering

`nuonidentity` stamps JWT-verified `nuon.org.id`, `nuon.app.id`, and `nuon.install.id` on each resource,
plus either `nuon.runner.id` or `nuon.collector.id`. The other identity key is absent. It first removes all
five keys from resource, scope, record, span/event/link, metric metadata/datapoint,
and exemplar attribute maps, matching case-insensitively and including fully underscored aliases such as
`nuon_org_id`. Other attributes, including unrelated Nuon IDs and `*.name` labels, are preserved.

Names are sender-supplied display/filtering labels: the relay neither looks them up nor verifies them against IDs.
Use verified IDs for authoritative filtering.

Restrict install telemetry by setting a YAML/JSON list of allowed JWT org IDs:

```sh
NUON_TELEMETRY_ALLOWED_ORG_IDS='["orgrok933tcyzji01s7us3aeo3", "org98e2wpzdxwoey393edtqj45"]'
```

The equivalent Collector setting is `processors.nuonidentity.allowed_org_ids`. Unset or empty allows all verified
orgs; entries must be nonblank, without surrounding whitespace. Matching is exact and uses the JWT org, not payload
IDs or names. Excluded requests are rejected before forwarding with a permanent error and may be discarded by the
sender; allowlist changes do not replay them. Environment telemetry and relay self-metrics are unaffected.

## Resource attributes

The standard resource processor overwrites resource-level `nuon.telemetry.source` with `install` or `environment`
to identify the ingestion path. Same-named attributes at other levels are untouched; query the resource attribute.

Customize `processors.resource/install.attributes` and `processors.resource/environment.attributes` in a mounted
Collector configuration or override file. Standard resource actions affect only resources: `insert` supplies a
missing value; `upsert` sets or replaces a value. For example:

```yaml
processors:
  resource/environment:
    attributes:
      - key: nuon.telemetry.source
        value: environment
        action: upsert
      - key: telemetry.dataset
        value: ${env:ENVIRONMENT_TELEMETRY_DATASET}
        action: upsert
      - key: deployment.environment.name
        value: production
        action: insert
```

Supply the override after the base configuration:

```sh
nuon-telemetry-relay \
  --config /etc/nuon-telemetry-relay/config.yaml \
  --config /etc/nuon-telemetry-relay/attributes.yaml
```

Override lists replace rather than append: retain the source-marker action and at least one action per processor.
`OTEL_RESOURCE_ATTRIBUTES` on the relay does not configure forwarded resource attributes.

Preserve producer identity (`service.*`, host/cloud/Kubernetes attributes) rather than assigning the relay's values
to every source. `nuon.control_plane.id` is optional and source-owned. Install resource actions run after
`nuonidentity`, so avoid overwriting the verified IDs. Operator configuration is trusted.

### Backend filtering

OTLP separates resource, scope, and record/datapoint attributes without defining cross-level precedence. Use
resource-aware filters or selective resource-to-label promotion, and check collision handling in flattening backends.

For example, in a Prometheus server accepting OTLP, merge these keys into its existing resource-promotion list:

```yaml
otlp:
  promote_resource_attributes:
    - nuon.telemetry.source
    - nuon.org.id
    - nuon.app.id
    - nuon.install.id
```

This belongs in the backend, not the relay. Retain existing promoted keys, budget for ID/name cardinality, and query
the backend's label spelling (for example, `nuon_telemetry_source` with underscore normalization). See
[Prometheus OTLP resource promotion](https://prometheus.io/docs/guides/opentelemetry/#promoting-resource-attributes).

## Relay self-observability

The relay exports process and pipeline metrics every 60 seconds through the environment listener, with a five-second
export budget and `service.name=nuon-telemetry-relay`. The Collector adds its version and a unique process instance
ID; environment resource actions also apply. Self-metrics contribute to environment traffic counters. Logs go to
stderr; internal traces are disabled. Retain independent health or scrape monitoring to detect relay/backend outages.
